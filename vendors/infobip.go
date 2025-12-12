package vendors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
	"github.com/Infinite-Locus-Product/thums_up_backend/pkg/circuitbreaker"
)

type InfobipClient struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
	cb      *circuitbreaker.CircuitBreaker
}

func InitInfobip() *InfobipClient {
	cfg := config.GetConfig()
	
	// Create circuit breaker for SMS service
	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		Name:             "infobip-sms",
		MaxRequests:      5,
		Interval:         60 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 5,
		SuccessThreshold: 2,
	})
	
	return &InfobipClient{
		BaseURL: cfg.InfobipConfig.BaseURL,
		APIKey:  cfg.InfobipConfig.APIKey,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cb: cb,
	}
}

func (ic *InfobipClient) SendSMS(ctx context.Context, to, message string) error {
	// Wrap the actual SMS sending with circuit breaker
	return ic.cb.Execute(ctx, func(ctx context.Context) error {
		return ic.sendSMSInternal(ctx, to, message)
	})
}

func (ic *InfobipClient) sendSMSInternal(ctx context.Context, to, message string) error {
	url := fmt.Sprintf("%s/sms/2/text/advanced", ic.BaseURL)

	payload := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"destinations": []map[string]string{
					{"to": to},
				},
				"text": message,
			},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal SMS payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "App "+ic.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ic.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.WithError(err).Warn("Failed to read response body")
		body = []byte{}
	}
	log.WithFields(log.Fields{
		"status":   resp.StatusCode,
		"response": string(body),
	}).Debug("SMS response")

	if resp.StatusCode >= 400 {
		return fmt.Errorf("SMS failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

