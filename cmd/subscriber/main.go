package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	pubsubpkg "cloud.google.com/go/pubsub"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
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

	ctx := context.Background()
	messageHandler := createMessageHandler()

	go func() {
		srv := setupHealthCheckServer(cfg.AppPort)
		log.Infof("Health check server starting on port %s", cfg.AppPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Warnf("Health check server error: %v", err)
		}
	}()

	go func() {
		if err := subscriber.Subscribe(ctx, cfg.PubSubConfig.SubscriptionID, messageHandler); err != nil {
			log.Fatalf("Failed to start subscriber: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down subscriber...")
	subscriber.Close()
	log.Info("Subscriber stopped")
}

func createMessageHandler() pubsub.MessageHandler {
	return func(ctx context.Context, msg *pubsubpkg.Message) error {
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
