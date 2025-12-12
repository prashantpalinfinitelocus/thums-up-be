package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/handlers"
	"github.com/Infinite-Locus-Product/thums_up_backend/middlewares"
	"github.com/Infinite-Locus-Product/thums_up_backend/repository"
	"github.com/Infinite-Locus-Product/thums_up_backend/services"
	"github.com/Infinite-Locus-Product/thums_up_backend/utils"
	"github.com/Infinite-Locus-Product/thums_up_backend/vendors"
)

var (
	db             *gorm.DB
	firebaseClient *vendors.FirebaseClient
	infobipClient  *vendors.InfobipClient
	pubsubClient   interface{}
	gcsService     utils.GCSService
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

	initVendors()
	initDatabase()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middlewares.CORSMiddleware())
	router.Use(middlewares.ErrorHandler())

	setupRoutes(router)

	port := cfg.AppPort
	if portFlag != "" {
		port = portFlag
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	go func() {
		log.Infof("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}

func initVendors() {
	cfg := config.GetConfig()

	infobipClient = vendors.InitInfobip()

	// firebaseClient = vendors.InitFirebase()

	// if cfg.PubSubConfig.ProjectID != "" {
	// 	pubsubClient = vendors.InitPubSub()
	// }

	var err error
	credPath := "./gcp-service-account.json"
	gcsService, err = utils.NewGCSServiceWithCredentials(cfg.GcsConfig.BucketName, cfg.GcsConfig.ProjectID, credPath)
	if err != nil {
		log.Warnf("Failed to initialize GCS service: %v", err)
	} else {
		log.Info("GCS service initialized successfully")
	}
}

func initDatabase() {
	db = vendors.InitDatabase()

	migrationRunner := utils.NewMigrationRunner(db, "migrations")
	if err := migrationRunner.Run(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	db.AutoMigrate(
		&entities.User{},
		&entities.OTPLog{},
		&entities.RefreshToken{},
		&entities.NotifyMe{},
		&entities.Address{},
		&entities.State{},
		&entities.City{},
		&entities.PinCode{},
		&entities.QuestionMaster{},
		&entities.ThunderSeat{},
		&entities.ThunderSeatWinner{},
	)

	log.Info("Database migrations completed")
}

func setupRoutes(router *gin.Engine) {
	txnManager := utils.NewTransactionManager(db)

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

	authService := services.NewAuthService(txnManager, userRepo, otpRepo, refreshTokenRepo, infobipClient)
	notifyMeService := services.NewNotifyMeService(txnManager, notifyMeRepo)
	userService := services.NewUserService(txnManager, userRepo, addressRepo, stateRepo, cityRepo, pinCodeRepo)
	questionService := services.NewQuestionService(txnManager, questionRepo)
	thunderSeatService := services.NewThunderSeatService(txnManager, thunderSeatRepo, questionRepo)
	winnerService := services.NewWinnerService(txnManager, winnerRepo, thunderSeatRepo)

	authHandler := handlers.NewAuthHandler(authService)
	notifyMeHandler := handlers.NewNotifyMeHandler(notifyMeService, nil)
	profileHandler := handlers.NewProfileHandler(userService)
	addressHandler := handlers.NewAddressHandler(userService)
	questionHandler := handlers.NewQuestionHandler(questionService)
	thunderSeatHandler := handlers.NewThunderSeatHandler(thunderSeatService)
	winnerHandler := handlers.NewWinnerHandler(winnerService)

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
			auth.POST("/send-otp", authHandler.SendOTP)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
			auth.POST("/signup", authHandler.SignUp)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		notifyMe := api.Group("/notify-me")
		{
			notifyMe.POST("/subscribe", notifyMeHandler.Subscribe)
			notifyMe.GET("/:phone", notifyMeHandler.GetSubscription)
		}

		profileGroup := api.Group("/profile")
		profileGroup.Use(middlewares.AuthMiddleware(db, userRepo))
		{
			profileGroup.GET("/", profileHandler.GetProfile)
			profileGroup.PATCH("/", profileHandler.UpdateProfile)

			profileGroup.POST("/address", addressHandler.AddAddress)
			profileGroup.GET("/address", addressHandler.GetAddresses)
			profileGroup.PUT("/address/:addressId", addressHandler.UpdateAddress)
			profileGroup.DELETE("/address/:addressId", addressHandler.DeleteAddress)
		}

		questions := api.Group("/questions")
		{
			questions.GET("/active", questionHandler.GetActiveQuestions)

			questionsAuth := questions.Group("")
			questionsAuth.Use(middlewares.AuthMiddleware(db, userRepo))
			{
				questionsAuth.POST("/", questionHandler.SubmitQuestion)
			}
		}

		thunderSeat := api.Group("/thunder-seat")
		{
			thunderSeat.GET("/current-week", thunderSeatHandler.GetCurrentWeek)

			thunderSeatAuth := thunderSeat.Group("")
			thunderSeatAuth.Use(middlewares.AuthMiddleware(db, userRepo))
			{
				thunderSeatAuth.GET("/submissions", thunderSeatHandler.GetUserSubmissions)
				thunderSeatAuth.POST("/", thunderSeatHandler.SubmitAnswer)
			}
		}

		winners := api.Group("/winners")
		{
			winners.GET("/", winnerHandler.GetAllWinners)
			winners.GET("/week/:weekNumber", winnerHandler.GetWinnersByWeek)
		}

		admin := api.Group("/admin")
		admin.Use(middlewares.APIKeyMiddleware())
		{
			admin.GET("/notify-me/unnotified", notifyMeHandler.GetAllUnnotified)
			admin.POST("/winners/select", winnerHandler.SelectWinners)
		}
	}
}
