package vendors

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
	"github.com/Infinite-Locus-Product/thums_up_backend/constants"
)

type FirebaseClient struct {
	messagingClient *messaging.Client
	app             *firebase.App
}

func InitFirebase() *FirebaseClient {
	cfg := config.GetConfig()

	firebaseConfig := &firebase.Config{
		ProjectID: cfg.GcsConfig.ProjectID,
	}
	app, err := firebase.NewApp(context.Background(), firebaseConfig)
	if err != nil {
		log.Errorf("Failed to initialize Firebase app: %v", err)
		return nil
	}

	messagingClient, err := app.Messaging(context.Background())
	if err != nil {
		log.Errorf("Failed to get messaging client: %v", err)
		return nil
	}

	log.Info("Firebase client initialized successfully")

	return &FirebaseClient{
		messagingClient: messagingClient,
		app:             app,
	}
}

func (fc *FirebaseClient) GetMessagingClient() *messaging.Client {
	return fc.messagingClient
}

func (fc *FirebaseClient) SendMessage(ctx context.Context, token string, title string, body string, data map[string]string) (string, error) {
	if data == nil {
		data = map[string]string{}
	}

	data["title"] = title
	data["body"] = body

	message := &messaging.Message{
		Token: token,
		Data:  data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Category: constants.NOTIFICATION_CATEGORY,
				},
			},
		},
	}

	response, err := fc.messagingClient.Send(ctx, message)
	if err != nil {
		return "", err
	}
	return response, nil
}

func (fc *FirebaseClient) SendBulkMessage(ctx context.Context, tokens []string, title string, body string, data map[string]string) (int, int, error) {
	if len(tokens) == 0 {
		return 0, 0, nil
	}

	if data == nil {
		data = map[string]string{}
	}

	data["title"] = title
	data["body"] = body

	const batchSize = 500
	totalSuccessCount := 0
	totalFailureCount := 0

	for i := 0; i < len(tokens); i += batchSize {
		end := i + batchSize
		if end > len(tokens) {
			end = len(tokens)
		}

		batchTokens := tokens[i:end]

		message := &messaging.MulticastMessage{
			Tokens: batchTokens,
			Data:   data,
			Android: &messaging.AndroidConfig{
				Priority: "high",
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Category: constants.NOTIFICATION_CATEGORY,
					},
				},
			},
		}

		batchResponse, err := fc.messagingClient.SendEachForMulticast(ctx, message)
		if err != nil {
			log.Errorf("Failed to send batch %d-%d: %v", i, end, err)
			continue
		}

		totalSuccessCount += batchResponse.SuccessCount
		totalFailureCount += batchResponse.FailureCount
	}

	return totalSuccessCount, totalFailureCount, nil
}
