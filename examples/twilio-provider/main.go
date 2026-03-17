// Example: sending SMS via Twilio.
//
// Set the following environment variables before running:
//
//	TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
//	TWILIO_AUTH_TOKEN=your_auth_token
//	TWILIO_FROM=+15551234567
//
// Run: go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/KARTIKrocks/gosms"
	"github.com/KARTIKrocks/gosms/twilio"
)

func main() {
	provider, err := twilio.NewProvider(twilio.Config{
		AccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
		AuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		From:       os.Getenv("TWILIO_FROM"),
	})
	if err != nil {
		log.Fatal(err)
	}

	client := gosms.NewClient(provider)
	ctx := context.Background()

	// Send a simple message.
	result, err := client.Send(ctx, "+15559876543", "Hello from gosms via Twilio!")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent: id=%s status=%s segments=%d\n", result.MessageID, result.Status, result.Segments)

	// Send a scheduled message (requires Twilio Messaging Service).
	msg := gosms.NewMessage("+15559876543", "Reminder: your appointment is tomorrow").
		WithSchedule(time.Now().Add(24 * time.Hour))

	result, err = client.SendMessage(ctx, msg)
	if err != nil {
		fmt.Printf("Scheduled send (expected to fail without messaging service): %v\n", err)
	} else {
		fmt.Printf("Scheduled: id=%s\n", result.MessageID)
	}

	// Check delivery status.
	status, err := client.GetStatus(ctx, result.MessageID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Status: %s\n", status.Status)
}
