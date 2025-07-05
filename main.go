package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"go-web-wa/pkg/config"
	"go-web-wa/pkg/discord"
	"go-web-wa/pkg/whatsapp"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting WhatsApp Profile Fetcher for: %s", cfg.TargetPhoneNumber)

	// Initialize Discord client
	discordClient := discord.NewWebhookClient(cfg.DiscordWebhookURL)

	// Initialize WhatsApp client
	waClient, err := whatsapp.NewClient(cfg.SessionFilePath)
	if err != nil {
		log.Printf("Failed to create WhatsApp client: %v", err)
		sendErrorToDiscord(discordClient, "WhatsApp Client Error", fmt.Sprintf("Failed to create WhatsApp client: %v", err))
		return
	}
	defer waClient.Close()

	// Check if paired/logged in
	if !waClient.IsLoggedIn() {
		log.Printf("WhatsApp client not logged in. Please run the pairing process first.")
		sendErrorToDiscord(discordClient, "Authentication Required", "WhatsApp client not logged in. Please run the pairing process first.")
		return
	}

	// Connect to WhatsApp
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log.Println("Connecting to WhatsApp...")
	if err := waClient.Connect(ctx); err != nil {
		log.Printf("Failed to connect to WhatsApp: %v", err)
		sendErrorToDiscord(discordClient, "Connection Error", fmt.Sprintf("Failed to connect to WhatsApp: %v", err))
		return
	}

	// Wait a moment for connection to stabilize
	time.Sleep(2 * time.Second)

	// Fetch profile picture
	log.Printf("Fetching profile picture for: %s", cfg.TargetPhoneNumber)
	imageData, err := waClient.GetProfilePicture(cfg.TargetPhoneNumber)
	if err != nil {
		log.Printf("Failed to fetch profile picture: %v", err)
		sendErrorToDiscord(discordClient, "Profile Picture Error", fmt.Sprintf("Failed to fetch profile picture for %s: %v", cfg.TargetPhoneNumber, err))
		return
	}

	// Generate filename
	filename := fmt.Sprintf("profile_%s_%s.jpg", cfg.TargetPhoneNumber, time.Now().Format("20060102_150405"))

	// Send image to Discord
	log.Println("Sending profile picture to Discord...")
	if err := discordClient.SendImageWithFile(imageData, filename, cfg.TargetPhoneNumber); err != nil {
		log.Printf("Failed to send image to Discord: %v", err)
		sendErrorToDiscord(discordClient, "Discord Error", fmt.Sprintf("Failed to send image to Discord: %v", err))
		return
	}

	// Send success message
	log.Println("Profile picture sent successfully!")
	discordClient.SendSuccessMessage(
		"Profile Picture Fetched",
		fmt.Sprintf("Successfully fetched and sent profile picture for %s", cfg.TargetPhoneNumber),
	)

	// Wait a moment for the message to be sent
	time.Sleep(2 * time.Second)

	// Disconnect from WhatsApp
	waClient.Disconnect()

	// Wait a moment for the message to be sent
	time.Sleep(2 * time.Second)

	log.Println("Task completed successfully!")
}

// sendErrorToDiscord sends an error message to Discord
func sendErrorToDiscord(client *discord.WebhookClient, title, message string) {
	if err := client.SendErrorMessage(title, message); err != nil {
		log.Printf("Failed to send error message to Discord: %v", err)
	}
}

// pairDevice handles the initial pairing process
func pairDevice() {
	// Load configuration
	sessionPath := os.Getenv("SESSION_FILE_PATH")
	if sessionPath == "" {
		sessionPath = "./sessions/"
	}

	// Initialize WhatsApp client
	waClient, err := whatsapp.NewClient(sessionPath)
	if err != nil {
		log.Fatalf("Failed to create WhatsApp client: %v", err)
	}
	defer waClient.Close()

	// Check if already paired
	if waClient.IsLoggedIn() {
		log.Println("Already logged in to WhatsApp")
		return
	}

	// Ask for pairing method
	fmt.Println("Choose pairing method:")
	fmt.Println("1. QR Code")
	fmt.Println("2. Phone Number")
	fmt.Print("Enter choice (1 or 2): ")

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		log.Println("Starting QR code pairing...")
		if err := waClient.PairQR(); err != nil {
			log.Fatalf("Failed to pair with QR code: %v", err)
		}
	case "2":
		fmt.Print("Enter your phone number (with country code, e.g., +1234567890): ")
		var phoneNumber string
		fmt.Scanln(&phoneNumber)

		// Clean phone number
		phoneNumber = strings.ReplaceAll(phoneNumber, "+", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, "-", "")
		phoneNumber = strings.ReplaceAll(phoneNumber, " ", "")

		log.Printf("Starting phone number pairing for: %s", phoneNumber)
		if err := waClient.PairPhone(phoneNumber); err != nil {
			log.Fatalf("Failed to pair with phone number: %v", err)
		}
	default:
		log.Fatalf("Invalid choice: %s", choice)
	}

	log.Println("Pairing completed successfully!")
}

// init function to check command line arguments
func init() {
	if len(os.Args) > 1 && os.Args[1] == "pair" {
		pairDevice()
		os.Exit(0)
	}
}
