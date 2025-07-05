package whatsapp

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"

	// Import SQLite driver
	_ "github.com/mattn/go-sqlite3"
)

// Client wraps whatsmeow client with additional functionality
type Client struct {
	client        *whatsmeow.Client
	store         *sqlstore.Container
	sessionPath   string
	isConnected   bool
	eventHandlers map[string]func(interface{})
}

// NewClient creates a new WhatsApp client
func NewClient(sessionPath string) (*Client, error) {
	// Ensure session directory exists
	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	// Create database path
	dbPath := filepath.Join(sessionPath, "whatsapp.db")

	// Create store
	dbLog := waLog.Stdout("Database", "ERROR", true)
	store, err := sqlstore.New(context.Background(), "sqlite3", "file:"+dbPath+"?_foreign_keys=on", dbLog)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	// Get device store
	deviceStore, err := store.GetFirstDevice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get device store: %w", err)
	}

	// Create client log
	clientLog := waLog.Stdout("Client", "ERROR", true)

	// Create whatsmeow client
	client := whatsmeow.NewClient(deviceStore, clientLog)

	waClient := &Client{
		client:        client,
		store:         store,
		sessionPath:   sessionPath,
		isConnected:   false,
		eventHandlers: make(map[string]func(interface{})),
	}

	// Add event handlers
	waClient.setupEventHandlers()

	return waClient, nil
}

// setupEventHandlers sets up event handlers for the client
func (c *Client) setupEventHandlers() {
	// TODO: Fix event handlers based on whatsmeow API
	// For now, we'll track connection status manually
	log.Println("Event handlers setup (simplified)")
}

// Connect connects to WhatsApp
func (c *Client) Connect(ctx context.Context) error {
	// Check if already logged in
	if c.client.Store.ID == nil {
		return fmt.Errorf("not logged in - please run pairing first")
	}

	// Connect
	err := c.client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Wait for connection with timeout
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("connection timeout")
		case <-ticker.C:
			if c.client.IsConnected() {
				log.Println("Successfully connected to WhatsApp")
				c.isConnected = true
				return nil
			}
		}
	}
}

// Disconnect disconnects from WhatsApp
func (c *Client) Disconnect() {
	if c.client != nil {
		c.client.Disconnect()
	}
}

// Close closes the client and store
func (c *Client) Close() error {
	if c.client != nil {
		c.client.Disconnect()
	}
	if c.store != nil {
		return c.store.Close()
	}
	return nil
}

// IsLoggedIn checks if the client is logged in
func (c *Client) IsLoggedIn() bool {
	return c.client.Store.ID != nil
}

// IsConnected checks if the client is connected
func (c *Client) IsConnected() bool {
	return c.client.IsConnected()
}

// PairPhone pairs the client with a phone number
func (c *Client) PairPhone(phoneNumber string) error {
	if c.client.Store.ID != nil {
		return fmt.Errorf("already logged in")
	}

	// Request pairing code
	code, err := c.client.PairPhone(context.Background(), phoneNumber, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
	if err != nil {
		return fmt.Errorf("failed to pair phone: %w", err)
	}

	fmt.Printf("Pairing code: %s\n", code)
	fmt.Println("Please enter this code in WhatsApp on your phone")

	// Wait for pairing to complete
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("pairing timeout")
		case <-ticker.C:
			if c.client.Store.ID != nil {
				log.Println("Successfully paired with WhatsApp")
				return nil
			}
		}
	}
}

// PairQR pairs the client using QR code
func (c *Client) PairQR() error {
	if c.client.Store.ID != nil {
		return fmt.Errorf("already logged in")
	}

	// Generate QR code
	qrChan, err := c.client.GetQRChannel(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get QR channel: %w", err)
	}

	// Handle QR events
	go func() {
		for evt := range qrChan {
			if evt.Event == "code" {
				fmt.Println("QR code:")
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Printf("QR event: %s\n", evt.Event)
			}
		}
	}()

	// Connect to start QR generation
	err = c.client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Wait for login
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("QR pairing timeout")
		case <-ticker.C:
			if c.client.Store.ID != nil {
				log.Println("Successfully paired with WhatsApp")
				return nil
			}
		}
	}
}

// GetProfilePicture fetches the profile picture of a phone number
func (c *Client) GetProfilePicture(phoneNumber string) ([]byte, error) {
	if !c.isConnected {
		return nil, fmt.Errorf("not connected to WhatsApp")
	}

	// Parse phone number to JID
	jid, err := c.parsePhoneNumber(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to parse phone number: %w", err)
	}

	// Get profile picture info
	profilePic, err := c.client.GetProfilePictureInfo(jid, &whatsmeow.GetProfilePictureParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get profile picture info: %w", err)
	}

	if profilePic == nil {
		return nil, fmt.Errorf("no profile picture found for %s", phoneNumber)
	}

	// Download the image
	imageData, err := c.downloadImage(profilePic.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download profile picture: %w", err)
	}

	return imageData, nil
}

// parsePhoneNumber parses a phone number to WhatsApp JID
func (c *Client) parsePhoneNumber(phoneNumber string) (types.JID, error) {
	// Remove any non-digit characters
	phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
	phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")

	// Create JID
	jid := types.NewJID(phoneNumber, types.DefaultUserServer)

	return jid, nil
}

// downloadImage downloads an image from URL
func (c *Client) downloadImage(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	// Read the image data
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return imageData, nil
}

// GetUserInfo gets user information for a phone number
func (c *Client) GetUserInfo(phoneNumber string) (*types.UserInfo, error) {
	if !c.isConnected {
		return nil, fmt.Errorf("not connected to WhatsApp")
	}

	jid, err := c.parsePhoneNumber(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to parse phone number: %w", err)
	}

	userInfo, err := c.client.GetUserInfo([]types.JID{jid})
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	if len(userInfo) == 0 {
		return nil, fmt.Errorf("no user info found for %s", phoneNumber)
	}

	info := userInfo[jid]
	return &info, nil
}
