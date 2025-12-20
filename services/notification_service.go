package services

import (
	"context"
	"encoding/json"

	"cloud.google.com/go/pubsub"
	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
	"github.com/Infinite-Locus-Product/thums_up_backend/constants"
	"github.com/Infinite-Locus-Product/thums_up_backend/vendors"
)

type NotificationService interface {
	SendNotification(ctx context.Context, token string, title string, body string, data map[string]string) error
	PublishNotifyMeMessage(ctx context.Context, phoneNumber string, email string) error
}

type notificationService struct {
	firebaseClient *vendors.FirebaseClient
	pubsubClient   *pubsub.Client
	cfg            *config.Config
}

func NewNotificationService(
	firebaseClient *vendors.FirebaseClient,
	pubsubClient *pubsub.Client,
) NotificationService {
	return &notificationService{
		firebaseClient: firebaseClient,
		pubsubClient:   pubsubClient,
		cfg:            config.GetConfig(),
	}
}

func (s *notificationService) SendNotification(ctx context.Context, token string, title string, body string, data map[string]string) error {
	if s.firebaseClient == nil {
		log.Warn("Firebase client not initialized")
		return nil
	}

	_, err := s.firebaseClient.SendMessage(ctx, token, title, body, data)
	if err != nil {
		log.WithError(err).Error("Failed to send notification")
		return err
	}

	return nil
}

func (s *notificationService) PublishNotifyMeMessage(ctx context.Context, phoneNumber string, email string) error {
	if s.pubsubClient == nil {
		log.Warn("PubSub client not initialized")
		return nil
	}

	message := map[string]string{
		"phone_number": phoneNumber,
		"email":        email,
		"type":         "notify_me",
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Error("Failed to marshal pubsub message")
		return err
	}

	topic := s.pubsubClient.Topic(s.cfg.PubSubConfig.TopicID)
	result := topic.Publish(ctx, &pubsub.Message{
		Data: data,
		Attributes: map[string]string{
			"type": constants.NOTIFY_ME_TOPIC,
		},
	})

	_, err = result.Get(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to publish message to PubSub")
		return err
	}

	log.Info("Notify me message published successfully")
	return nil
}
