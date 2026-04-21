// Example: handling delivery status webhooks from Twilio, Vonage, and MSG91.
//
// Run: go run .
// Then POST to /webhook/twilio, /webhook/vonage, or /webhook/msg91.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/KARTIKrocks/gosms"
	"github.com/KARTIKrocks/gosms/msg91"
	"github.com/KARTIKrocks/gosms/twilio"
	"github.com/KARTIKrocks/gosms/vonage"
)

func main() {
	// Register parsers generically via gosms.WebhookParser.
	parsers := map[string]gosms.WebhookParser{
		"/webhook/twilio": twilio.ParseWebhook,
		"/webhook/vonage": vonage.ParseWebhook,
		"/webhook/msg91":  msg91.ParseWebhook,
	}

	for path, parse := range parsers {
		http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			status, err := parse(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			fmt.Printf("[%s] Message %s → %s\n", path, status.MessageID, status.Status)

			if status.Status.IsFinal() {
				if status.Status.IsSuccess() {
					fmt.Printf("  Delivered at %v\n", status.UpdatedAt)
				} else {
					fmt.Printf("  Failed: %s (code: %s)\n", status.ErrorMessage, status.ErrorCode)
				}
			}

			w.WriteHeader(http.StatusOK)
		})
	}

	fmt.Println("Webhook server listening on :8080")
	fmt.Println("  POST /webhook/twilio  — Twilio status callbacks")
	fmt.Println("  POST /webhook/vonage  — Vonage delivery receipts")
	fmt.Println("  POST /webhook/msg91   — MSG91 delivery reports")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
