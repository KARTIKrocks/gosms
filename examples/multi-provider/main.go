// Example: using MultiProvider for fallback and round-robin strategies.
//
// This uses mock providers to demonstrate the behavior without real credentials.
//
// Run: go run .
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/KARTIKrocks/gosms"
)

func main() {
	ctx := context.Background()

	// Create two mock providers to simulate multiple backends.
	primary := gosms.NewMockProvider()
	fallback := gosms.NewMockProvider()

	// --- Fallback strategy ---
	// Tries providers in order; if the first fails, falls back to the next.
	multi := gosms.NewMultiProvider(primary, fallback)
	client := gosms.NewClient(multi).WithDefaultFrom("+15551234567")

	result, err := client.Send(ctx, "+15559876543", "Fallback test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Fallback: sent via provider, id=%s\n", result.MessageID)

	// Simulate primary failure — message routes to fallback.
	primary.WithSendError(errors.New("primary down"))
	result, err = client.Send(ctx, "+15559876543", "Primary is down")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Fallback: primary down, sent via fallback, id=%s\n", result.MessageID)
	primary.WithSendError(nil) // restore

	// --- Round-robin strategy ---
	// Rotates across providers for load distribution.
	rr := gosms.NewMultiProvider(primary, fallback).
		WithStrategy(gosms.StrategyRoundRobin)
	rrClient := gosms.NewClient(rr).WithDefaultFrom("+15551234567")

	for i := range 4 {
		result, err := rrClient.Send(ctx, "+15559876543", fmt.Sprintf("Round-robin message %d", i+1))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Round-robin #%d: id=%s\n", i+1, result.MessageID)
	}

	fmt.Printf("\nPrimary sent: %d, Fallback sent: %d\n",
		primary.MessageCount(), fallback.MessageCount())
}
