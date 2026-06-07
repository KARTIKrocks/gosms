// Example: basic usage of gosms with the mock provider.
//
// This demonstrates the core API without requiring real provider credentials.
//
// Run: go run .
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/KARTIKrocks/gosms"
)

func main() {
	ctx := context.Background()

	// Create a mock provider (no credentials needed).
	mock := gosms.NewMockProvider()

	// Create a client with a default sender.
	client := gosms.NewClient(mock).WithDefaultFrom("+15551234567")

	// --- Simple send ---
	result, err := client.Send(ctx, "+15559876543", "Hello from gosms!")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent: id=%s status=%s\n", result.MessageID, result.Status)

	// --- Message builder ---
	msg := gosms.NewMessage("+15559876543", "Your order has shipped!").
		WithFrom("+15551234567").
		WithReference("order-456").
		WithValidity(1*time.Hour).
		WithMetadata("order_id", "456")

	result, err = client.SendMessage(ctx, msg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent: id=%s ref=%s\n", result.MessageID, msg.Reference)

	// --- Bulk send with Batch ---
	batch := gosms.NewBatch()
	batch.AddNew("+15551111111", "Batch message 1")
	batch.AddNew("+15552222222", "Batch message 2")
	batch.AddNew("+15553333333", "Batch message 3")

	results, err := batch.Send(ctx, client)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range results {
		fmt.Printf("Batch: to=%s status=%s\n", r.To, r.Status)
	}

	// --- SendToMany (same message, multiple recipients) ---
	results, err = gosms.SendToMany(ctx, client,
		"Flash sale! 50% off today!",
		"+15554444444", "+15555555555",
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("SendToMany: sent %d messages\n", len(results))

	// --- Delivery status ---
	status, err := client.GetStatus(ctx, result.MessageID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Status: %s (final=%t)\n", status.Status, status.Status.IsFinal())

	// --- Message templates ---
	otp := gosms.OTPMessage("+15559876543", "482910", "MyApp")
	fmt.Printf("OTP body: %s\n", otp.Body)

	alert := gosms.AlertMessage("+15559876543", "CRITICAL", "CPU usage at 98%")
	fmt.Printf("Alert body: %s\n", alert.Body)

	notif := gosms.NotificationMessage("+15559876543", "Shipping", "Your package is out for delivery")
	fmt.Printf("Notification body: %s\n", notif.Body)

	fmt.Printf("\nTotal messages sent: %d\n", mock.MessageCount())
}
