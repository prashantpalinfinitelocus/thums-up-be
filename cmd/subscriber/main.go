package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	pubsubpkg "cloud.google.com/go/pubsub"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
	"github.com/Infinite-Locus-Product/thums_up_backend/constants"
	"github.com/Infinite-Locus-Product/thums_up_backend/entities"
	"github.com/Infinite-Locus-Product/thums_up_backend/pubsub"
	"github.com/Infinite-Locus-Product/thums_up_backend/vendors"
)

var (
	db             *gorm.DB
	firebaseClient *vendors.FirebaseClient
)

func main() {
	initLogging()
	startSubscriber()
}

func initLogging() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func startSubscriber() {
	cfg := config.GetConfig()

	db = vendors.InitDatabase()
	db.AutoMigrate(
		&entities.User{},
		&entities.NotifyMe{},
	)

	firebaseClient = vendors.InitFirebase()
	pubsubClient := vendors.InitPubSub()

	subscriber := pubsub.NewSubscriber(pubsubClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messageHandler := createMessageHandler()

	srv := setupHealthCheckServer(cfg.AppPort)
	go func() {
		log.Infof("Health check server starting on port %s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Warnf("Health check server error: %v", err)
		}
	}()

	go func() {
		if err := subscriber.Subscribe(ctx, cfg.PubSubConfig.SubscriptionID, messageHandler); err != nil {
			log.WithError(err).Error("Subscriber error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down subscriber...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Warn("Health check server shutdown error")
	}

	subscriber.Close()
	log.Info("Subscriber stopped")
}

func createMessageHandler() pubsub.MessageHandler {
	return func(ctx context.Context, msg *pubsubpkg.Message) error {
		// Add timeout for message processing
		handlerCtx, cancel := context.WithTimeout(ctx, constants.MESSAGE_HANDLER_TIMEOUT)
		defer cancel()

		log.WithFields(log.Fields{
			"message_id":   msg.ID,
			"publish_time": msg.PublishTime,
			"attributes":   msg.Attributes,
		}).Info("Message received from PubSub")

		var data map[string]interface{}
		if err := pubsub.UnmarshalMessage(msg, &data); err != nil {
			log.WithError(err).Error("Failed to unmarshal message")
			return err
		}

		log.WithField("data", data).Debug("Message data")

		// Use handlerCtx if you need to do any work
		_ = handlerCtx

		return nil
	}
}

func setupHealthCheckServer(port string) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"thums-up-subscriber"}`)
	})

	return &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}
}
