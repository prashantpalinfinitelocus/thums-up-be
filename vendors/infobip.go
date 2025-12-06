package vendors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Infinite-Locus-Product/thums_up_backend/config"
)

type InfobipClient struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

func InitInfobip() *InfobipClient {
	cfg := config.GetConfig()
	return &InfobipClient{
		BaseURL: cfg.InfobipConfig.BaseURL,
		APIKey:  cfg.InfobipConfig.APIKey,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (ic *InfobipClient) SendSMS(to, message string) error {
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

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
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

	body, _ := io.ReadAll(resp.Body)
	log.WithFields(log.Fields{
		"status":   resp.StatusCode,
		"response": string(body),
	}).Debug("SMS response")

	if resp.StatusCode >= 400 {
		return fmt.Errorf("SMS failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
