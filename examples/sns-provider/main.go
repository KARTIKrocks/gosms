// Example: sending SMS via AWS SNS.
//
// Set the following environment variables before running:
//
//	AWS_REGION=us-east-1
//	AWS_ACCESS_KEY_ID=your_access_key
//	AWS_SECRET_ACCESS_KEY=your_secret_key
//
// Run: go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/KARTIKrocks/gosms"
	"github.com/KARTIKrocks/gosms/sns"
)

func main() {
	ctx := context.Background()

	config := sns.DefaultConfig()
	config.Region = os.Getenv("AWS_REGION")
	config.AccessKeyID = os.Getenv("AWS_ACCESS_KEY_ID")
	config.SecretAccessKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	config.SenderID = "MyApp"
	config.SMSType = sns.SMSTransactional

	provider, err := sns.NewProvider(ctx, config)
	if err != nil {
		log.Fatal(err)
	}

	client := gosms.NewClient(provider)

	// Send a transactional SMS.
	result, err := client.Send(ctx, "+15559876543", "Your verification code is 482910")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent: id=%s status=%s\n", result.MessageID, result.Status)

	// Check opt-out status.
	optedOut, err := provider.CheckIfPhoneNumberIsOptedOut(ctx, "+15559876543")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Opted out: %t\n", optedOut)
}
