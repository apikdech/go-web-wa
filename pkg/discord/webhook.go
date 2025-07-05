package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

// WebhookClient handles Discord webhook operations
type WebhookClient struct {
	webhookURL string
	httpClient *http.Client
}

// NewWebhookClient creates a new Discord webhook client
func NewWebhookClient(webhookURL string) *WebhookClient {
	return &WebhookClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// MessagePayload represents a Discord webhook message payload
type MessagePayload struct {
	Content string  `json:"content,omitempty"`
	Embeds  []Embed `json:"embeds,omitempty"`
}

// Embed represents a Discord embed
type Embed struct {
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	Color       int     `json:"color,omitempty"`
	Timestamp   string  `json:"timestamp,omitempty"`
	Footer      *Footer `json:"footer,omitempty"`
	Image       *Image  `json:"image,omitempty"`
}

// Footer represents a Discord embed footer
type Footer struct {
	Text string `json:"text,omitempty"`
}

// Image represents a Discord embed image
type Image struct {
	URL string `json:"url,omitempty"`
}

// SendMessage sends a text message to Discord
func (c *WebhookClient) SendMessage(message string) error {
	payload := MessagePayload{
		Content: message,
	}

	return c.sendPayload(payload)
}

// SendErrorMessage sends an error message with embed styling
func (c *WebhookClient) SendErrorMessage(title, description string) error {
	payload := MessagePayload{
		Embeds: []Embed{
			{
				Title:       title,
				Description: description,
				Color:       0xFF0000, // Red color for errors
				Timestamp:   time.Now().Format(time.RFC3339),
				Footer: &Footer{
					Text: "WhatsApp Profile Fetcher",
				},
			},
		},
	}

	return c.sendPayload(payload)
}

// SendSuccessMessage sends a success message with embed styling
func (c *WebhookClient) SendSuccessMessage(title, description string) error {
	payload := MessagePayload{
		Embeds: []Embed{
			{
				Title:       title,
				Description: description,
				Color:       0x00FF00, // Green color for success
				Timestamp:   time.Now().Format(time.RFC3339),
				Footer: &Footer{
					Text: "WhatsApp Profile Fetcher",
				},
			},
		},
	}

	return c.sendPayload(payload)
}

// SendImageWithFile sends an image file to Discord
func (c *WebhookClient) SendImageWithFile(imageData []byte, filename, phoneNumber string) error {
	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add the image file
	fileWriter, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = fileWriter.Write(imageData)
	if err != nil {
		return fmt.Errorf("failed to write image data: %w", err)
	}

	// Add the payload data
	payloadWriter, err := writer.CreateFormField("payload_json")
	if err != nil {
		return fmt.Errorf("failed to create payload field: %w", err)
	}

	payload := MessagePayload{
		Embeds: []Embed{
			{
				Title:       "WhatsApp Profile Image",
				Description: fmt.Sprintf("Profile image for: %s", phoneNumber),
				Color:       0x0099FF, // Blue color for info
				Timestamp:   time.Now().Format(time.RFC3339),
				Footer: &Footer{
					Text: "WhatsApp Profile Fetcher",
				},
			},
		},
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	_, err = payloadWriter.Write(payloadJSON)
	if err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	// Send the request
	req, err := http.NewRequest("POST", c.webhookURL, &buf)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord webhook returned error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

// sendPayload sends a JSON payload to Discord
func (c *WebhookClient) sendPayload(payload MessagePayload) error {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.webhookURL, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord webhook returned error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}
