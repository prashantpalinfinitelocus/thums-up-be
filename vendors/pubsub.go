package vendors

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/Infinite-Locus-Product/thums_up_backend/config"
	log "github.com/sirupsen/logrus"
)

func InitPubSub() *pubsub.Client {
	cfg := config.GetConfig()
	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, cfg.PubSubConfig.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create pubsub client: %v", err)
	}
	log.Info("PubSub client created successfully")
	return client
}
