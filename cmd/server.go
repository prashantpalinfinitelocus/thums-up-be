package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
	"github.com/Infinite-Locus-Product/thums_up_backend/constants"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/pkg/queue"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
	"github.com/Infinite-Locus-Product/thums_up_backend/vendors"
)

var (
	portFlag string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func init() {
	serverCmd.Flags().StringVarP(&portFlag, "port", "p", "", "Port to run the server on (default from .env)")
	rootCmd.AddCommand(serverCmd)
}

func startServer() {
	cfg := config.GetConfig()

	srv := NewServer(cfg)
	defer srv.Cleanup()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := srv.setupRouter()

	port := cfg.AppPort
	if portFlag != "" {
		port = portFlag
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	go func() {
		log.Infof("Server starting on port %s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	srv.waitForShutdown(httpServer)

	log.Info("Server exited successfully")
}

func NewServer(cfg *config.Config) *Server {
	srv := &Server{
		cfg: cfg,
	}

	srv.initVendors()
	srv.initDatabase()
	srv.initWorkerPool()
	srv.initRepositories()
	srv.initHandlers()

	log.Info("Server initialized successfully")
	return srv
}

func (s *Server) initVendors() {
	s.infobipClient = vendors.InitInfobip()
	log.Info("Infobip client initialized")

	if err := s.initGCSService(); err != nil {
		log.Fatalf("Failed to initialize GCS service (required): %v", err)
	}
}

func (s *Server) initGCSService() error {
	if s.cfg.GcsConfig.BucketName == "" {
		return fmt.Errorf("GCP_BUCKET_NAME is not set in configuration")
	}
	if s.cfg.GcsConfig.ProjectID == "" {
		return fmt.Errorf("GCP_PROJECT_ID is not set in configuration")
	}

	credPath := "./gcp-service-account.json"
	log.Infof("Initializing GCS service with bucket: %s, project: %s", s.cfg.GcsConfig.BucketName, s.cfg.GcsConfig.ProjectID)

	gcsService, err := utils.NewGCSServiceWithCredentials(
		s.cfg.GcsConfig.BucketName,
		s.cfg.GcsConfig.ProjectID,
		credPath,
	)
	if err != nil {
		return fmt.Errorf("GCS initialization failed: %w", err)
	}

	s.gcsService = gcsService
	log.Info("GCS service initialized successfully")
	return nil
}

func (s *Server) initDatabase() {
	s.db = vendors.InitDatabase()

	if err := utils.RunDBMigrations(s.db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	log.Info("Database initialized and migrations completed")
}

func (s *Server) initWorkerPool() {
	s.workerPool = queue.NewWorkerPool(
		constants.WORKER_POOL_SIZE,
		constants.TASK_QUEUE_SIZE,
	)
	log.Info("Worker pool initialized successfully")
}

func (s *Server) initRepositories() {
	s.repositories = &Repositories{
		user:                   repository.NewUserRepository(),
		otp:                    repository.NewOTPRepository(),
		refreshToken:           repository.NewRefreshTokenRepository(),
		address:                repository.NewGormRepository[entities.Address](),
		avatar:                 repository.NewGormRepository[entities.Avatar](),
		state:                  repository.NewStateRepository(),
		city:                   repository.NewCityRepository(),
		pinCode:                repository.NewPinCodeRepository(),
		question:               repository.NewQuestionRepository(),
		questionMasterLanguage: repository.NewQuestionMasterLanguageRepository(s.db),
		optionMaster:           repository.NewOptionMasterRepository(s.db),
		optionMasterLanguage:   repository.NewOptionMasterLanguageRepository(s.db),
		userQuestionAnswer:     repository.NewUserQuestionAnswerRepository(s.db),
		thunderSeat:            repository.NewThunderSeatRepository(),
		winner:                 repository.NewWinnerRepository(),
		contestWeek:            repository.NewContestWeekRepository(),
		userAadharCard:         repository.NewUserAadharCardRepository(),
		userAdditionalInfo:     repository.NewUserAdditionalInfoRepository(),
		loginCount:             repository.NewLoginCountRepository(),
	}
	log.Debug("All repositories initialized")
}

func (s *Server) initHandlers() {
	txnManager := utils.NewTransactionManager(s.db)

	authService := services.NewAuthService(
		txnManager,
		s.repositories.user,
		s.repositories.otp,
		s.repositories.refreshToken,
		s.repositories.loginCount,
		s.infobipClient,
	)

	userService := services.NewUserService(
		txnManager,
		s.repositories.user,
		s.repositories.address,
		s.repositories.state,
		s.repositories.city,
		s.repositories.pinCode,
		s.repositories.avatar,
		s.gcsService,
		s.repositories.userQuestionAnswer,
		s.repositories.question,
		s.repositories.questionMasterLanguage,
		s.repositories.optionMaster,
		s.repositories.optionMasterLanguage,
	)

	avatarService := services.NewAvatarService(
		txnManager,
		s.repositories.avatar,
		s.gcsService,
	)

	questionService := services.NewQuestionService(
		txnManager,
		s.repositories.question,
		s.repositories.userQuestionAnswer,
		s.repositories.optionMaster,
	)

	contestWeekService := services.NewContestWeekService(
		txnManager,
		s.repositories.contestWeek,
	)

	thunderSeatService := services.NewThunderSeatService(
		txnManager,
		s.repositories.thunderSeat,
		s.repositories.contestWeek,
		s.repositories.user,
		s.gcsService,
	)

	winnerService := services.NewWinnerService(
		txnManager,
		s.repositories.winner,
		s.repositories.thunderSeat,
		s.repositories.contestWeek,
		s.repositories.user,
		s.repositories.userAadharCard,
		s.repositories.userAdditionalInfo,
		s.gcsService,
	)

	websiteStatusService := services.NewWebsiteStatusService(s.db, s.repositories.winner, s.repositories.contestWeek)

	stateService := services.NewStateService(s.db, s.repositories.state)

	s.handlers = &Handlers{
		auth:          handlers.NewAuthHandler(authService),
		profile:       handlers.NewProfileHandler(userService),
		address:       handlers.NewAddressHandler(userService),
		avatar:        handlers.NewAvatarHandler(avatarService),
		question:      handlers.NewQuestionHandler(questionService, userService),
		thunderSeat:   handlers.NewThunderSeatHandler(thunderSeatService),
		winner:        handlers.NewWinnerHandler(winnerService, s.gcsService),
		contestWeek:   handlers.NewContestWeekHandler(contestWeekService),
		websiteStatus: handlers.NewWebsiteStatusHandler(websiteStatusService),
		state:         handlers.NewStateHandler(stateService),
	}

	log.Debug("All handlers initialized")
}

func (s *Server) waitForShutdown(httpServer *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutdown signal received, initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		constants.GRACEFUL_SHUTDOWN_TIMEOUT,
	)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}
}

func (s *Server) Cleanup() {
	log.Info("Cleaning up server resources...")

	if s.workerPool != nil {
		s.workerPool.Shutdown()
		log.Debug("Worker pool shut down")
	}

	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Errorf("Error closing database: %v", err)
			} else {
				log.Debug("Database connection closed")
			}
		}
	}

	log.Info("Cleanup complete")
}
