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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
	"github.com/Infinite-Locus-Product/thums_up_backend/constants"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
	"github.com/Infinite-Locus-Product/thums_up_backend/pkg/queue"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
	"github.com/Infinite-Locus-Product/thums_up_backend/vendors"
)

var (
	portFlag string
)

// Server encapsulates all dependencies
type Server struct {
	db             *gorm.DB
	cfg            *config.Config
	firebaseClient *vendors.FirebaseClient
	infobipClient  *vendors.InfobipClient
	gcsService     utils.GCSService
	workerPool     *queue.WorkerPool
	handlers       *Handlers
}

// Handlers contains all HTTP handlers
type Handlers struct {
	auth        *handlers.AuthHandler
	notifyMe    *handlers.NotifyMeHandler
	profile     *handlers.ProfileHandler
	address     *handlers.AddressHandler
	question    *handlers.QuestionHandler
	thunderSeat *handlers.ThunderSeatHandler
	winner      *handlers.WinnerHandler
}

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

	// Initialize server
	srv := NewServer(cfg)
	defer srv.Cleanup()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.ErrorHandler())

	srv.setupRoutes(router)

	port := cfg.AppPort
	if portFlag != "" {
		port = portFlag
	}

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Start server
	go func() {
		log.Infof("Server starting on port %s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), constants.GRACEFUL_SHUTDOWN_TIMEOUT)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}

// NewServer creates and initializes a new server instance
func NewServer(cfg *config.Config) *Server {
	srv := &Server{
		cfg: cfg,
	}

	srv.initVendors()
	srv.initDatabase()
	srv.initWorkerPool()
	srv.initHandlers()

	return srv
}

func (s *Server) initVendors() {
	s.infobipClient = vendors.InitInfobip()

	// Initialize GCS Service
	var err error
	credPath := "./gcp-service-account.json"
	s.gcsService, err = utils.NewGCSServiceWithCredentials(
		s.cfg.GcsConfig.BucketName,
		s.cfg.GcsConfig.ProjectID,
		credPath,
	)
	if err != nil {
		log.Warnf("Failed to initialize GCS service: %v", err)
	} else {
		log.Info("GCS service initialized successfully")
	}

	// Initialize Firebase (optional)
	// s.firebaseClient = vendors.InitFirebase()
}

func (s *Server) initDatabase() {
	s.db = vendors.InitDatabase()

	if err := utils.RunDBMigrations(s.db); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}
}

func (s *Server) initWorkerPool() {
	s.workerPool = queue.NewWorkerPool(constants.WORKER_POOL_SIZE, constants.TASK_QUEUE_SIZE)
	log.Info("Worker pool initialized successfully")
}

func (s *Server) initHandlers() {
	txnManager := utils.NewTransactionManager(s.db)

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	otpRepo := repository.NewOTPRepository()
	refreshTokenRepo := repository.NewRefreshTokenRepository()
	notifyMeRepo := repository.NewNotifyMeRepository()
	addressRepo := repository.NewGormRepository[entities.Address]()
	stateRepo := repository.NewStateRepository()
	cityRepo := repository.NewCityRepository()
	pinCodeRepo := repository.NewPinCodeRepository()
	questionRepo := repository.NewQuestionRepository()
	thunderSeatRepo := repository.NewThunderSeatRepository()
	winnerRepo := repository.NewWinnerRepository()

	// Initialize services
	authService := services.NewAuthService(txnManager, userRepo, otpRepo, refreshTokenRepo, s.infobipClient)
	notifyMeService := services.NewNotifyMeService(txnManager, notifyMeRepo)
	userService := services.NewUserService(txnManager, userRepo, addressRepo, stateRepo, cityRepo, pinCodeRepo)
	questionService := services.NewQuestionService(txnManager, questionRepo)
	thunderSeatService := services.NewThunderSeatService(txnManager, thunderSeatRepo, questionRepo)
	winnerService := services.NewWinnerService(txnManager, winnerRepo, thunderSeatRepo)

	// Initialize handlers
	s.handlers = &Handlers{
		auth:        handlers.NewAuthHandler(authService),
		notifyMe:    handlers.NewNotifyMeHandler(notifyMeService, nil, s.workerPool),
		profile:     handlers.NewProfileHandler(userService),
		address:     handlers.NewAddressHandler(userService),
		question:    handlers.NewQuestionHandler(questionService),
		thunderSeat: handlers.NewThunderSeatHandler(thunderSeatService),
		winner:      handlers.NewWinnerHandler(winnerService),
	}
}

func (s *Server) setupRoutes(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "thums-up-backend",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/send-otp", s.handlers.auth.SendOTP)
			auth.POST("/verify-otp", s.handlers.auth.VerifyOTP)
			auth.POST("/signup", s.handlers.auth.SignUp)
			auth.POST("/refresh", s.handlers.auth.RefreshToken)
		}

		notifyMe := api.Group("/notify-me")
		{
			notifyMe.POST("/subscribe", s.handlers.notifyMe.Subscribe)
			notifyMe.GET("/:phone", s.handlers.notifyMe.GetSubscription)
		}

		userRepo := repository.NewUserRepository()
		profileGroup := api.Group("/profile")
		profileGroup.Use(middlewares.AuthMiddleware(s.db, userRepo))
		{
			profileGroup.GET("/", s.handlers.profile.GetProfile)
			profileGroup.PATCH("/", s.handlers.profile.UpdateProfile)

			profileGroup.POST("/address", s.handlers.address.AddAddress)
			profileGroup.GET("/address", s.handlers.address.GetAddresses)
			profileGroup.PUT("/address/:addressId", s.handlers.address.UpdateAddress)
			profileGroup.DELETE("/address/:addressId", s.handlers.address.DeleteAddress)
		}

		questions := api.Group("/questions")
		{
			questions.GET("/active", s.handlers.question.GetActiveQuestions)

			questionsAuth := questions.Group("")
			questionsAuth.Use(middlewares.AuthMiddleware(s.db, userRepo))
			{
				questionsAuth.POST("/", s.handlers.question.SubmitQuestion)
			}
		}

		thunderSeat := api.Group("/thunder-seat")
		{
			thunderSeat.GET("/current-week", s.handlers.thunderSeat.GetCurrentWeek)

			thunderSeatAuth := thunderSeat.Group("")
			thunderSeatAuth.Use(middlewares.AuthMiddleware(s.db, userRepo))
			{
				thunderSeatAuth.GET("/submissions", s.handlers.thunderSeat.GetUserSubmissions)
				thunderSeatAuth.POST("/", s.handlers.thunderSeat.SubmitAnswer)
			}
		}

		winners := api.Group("/winners")
		{
			winners.GET("/", s.handlers.winner.GetAllWinners)
			winners.GET("/week/:weekNumber", s.handlers.winner.GetWinnersByWeek)
		}

		admin := api.Group("/admin")
		admin.Use(middlewares.APIKeyMiddleware())
		{
			admin.GET("/notify-me/unnotified", s.handlers.notifyMe.GetAllUnnotified)
			admin.POST("/winners/select", s.handlers.winner.SelectWinners)
		}
	}
}

// Cleanup gracefully shuts down all resources
func (s *Server) Cleanup() {
	log.Info("Cleaning up server resources...")

	if s.workerPool != nil {
		s.workerPool.Shutdown()
	}

	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	log.Info("Cleanup complete")
}
