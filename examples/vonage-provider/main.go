// Example: sending SMS via Vonage (Nexmo).
//
// Set the following environment variables before running:
//
//	VONAGE_API_KEY=your_api_key
//	VONAGE_API_SECRET=your_api_secret
//	VONAGE_FROM=MyApp
//
// Run: go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/KARTIKrocks/gosms"
	"github.com/KARTIKrocks/gosms/vonage"
)

func main() {
	provider, err := vonage.NewProvider(vonage.Config{
		APIKey:    os.Getenv("VONAGE_API_KEY"),
		APISecret: os.Getenv("VONAGE_API_SECRET"),
		From:      os.Getenv("VONAGE_FROM"),
	})
	if err != nil {
		log.Fatal(err)
	}

	client := gosms.NewClient(provider)
	ctx := context.Background()

	// Send a simple message.
	result, err := client.Send(ctx, "+15559876543", "Hello from gosms via Vonage!")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent: id=%s status=%s cost=%s %s\n",
		result.MessageID, result.Status, result.Cost, result.Currency)
}
