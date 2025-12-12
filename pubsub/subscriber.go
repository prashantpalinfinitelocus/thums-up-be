package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/constants"
)

type Subscriber struct {
	client *pubsub.Client
	subs   map[string]*pubsub.Subscription
	mu     sync.Mutex
}

type MessageHandler func(ctx context.Context, msg *pubsub.Message) error

func NewSubscriber(client *pubsub.Client) *Subscriber {
	return &Subscriber{
		client: client,
		subs:   make(map[string]*pubsub.Subscription),
	}
}

func (s *Subscriber) Subscribe(ctx context.Context, subscriptionID string, handler MessageHandler) error {
	s.mu.Lock()
	sub := s.client.Subscription(subscriptionID)
	s.subs[subscriptionID] = sub
	s.mu.Unlock()

	log.Infof("Starting to receive messages from subscription: %s", subscriptionID)

	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		startTime := time.Now()

		log.WithFields(log.Fields{
			"message_id":   msg.ID,
			"publish_time": msg.PublishTime,
			"attributes":   msg.Attributes,
		}).Info("Processing PubSub message")

		if err := handler(ctx, msg); err != nil {
			log.WithFields(log.Fields{
				"message_id": msg.ID,
				"error":      err.Error(),
				"duration":   time.Since(startTime),
			}).Error("Handler failed to process message")
			msg.Nack()
			return
		}

		msg.Ack()
		log.WithFields(log.Fields{
			"message_id": msg.ID,
			"duration":   time.Since(startTime),
		}).Info("Message processed successfully")
	})

	if err != nil {
		log.WithError(err).Error("Error receiving messages from subscription")
		return err
	}

	return nil
}

func (s *Subscriber) CreateSubscriptionIfNotExists(ctx context.Context, subscriptionID, topicID string) error {
	sub := s.client.Subscription(subscriptionID)
	exists, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check subscription existence: %w", err)
	}

	if exists {
		log.WithField("subscription", subscriptionID).Info("Subscription already exists")
		return nil
	}

	topic := s.client.Topic(topicID)
	_, err = s.client.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
		Topic:       topic,
		AckDeadline: constants.PUBSUB_ACK_DEADLINE,
	})
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}

	log.WithFields(log.Fields{
		"subscription": subscriptionID,
		"topic":        topicID,
	}).Info("Subscription created successfully")

	return nil
}

func (s *Subscriber) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id := range s.subs {
		log.Infof("Stopped subscription: %s", id)
	}

	if s.client != nil {
		if err := s.client.Close(); err != nil {
			log.WithError(err).Error("Failed to close PubSub client")
		} else {
			log.Info("PubSub client closed successfully")
		}
	}
}

func UnmarshalMessage(msg *pubsub.Message, target interface{}) error {
	if err := json.Unmarshal(msg.Data, target); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return nil
}
