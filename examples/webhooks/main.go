// Example: handling delivery status webhooks from Twilio and Vonage.
//
// Run: go run .
// Then POST to http://localhost:8080/webhook/twilio or /webhook/vonage.
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/KARTIKrocks/gosms/twilio"
	"github.com/KARTIKrocks/gosms/vonage"
)

func main() {
	// Twilio delivery status webhook.
	http.HandleFunc("/webhook/twilio", func(w http.ResponseWriter, r *http.Request) {
		status, err := twilio.ParseWebhook(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("[Twilio] Message %s → %s\n", status.MessageID, status.Status)

		if status.Status.IsFinal() {
			if status.Status.IsSuccess() {
				fmt.Printf("  Delivered at %v\n", status.UpdatedAt)
			} else {
				fmt.Printf("  Failed: %s (code: %s)\n", status.ErrorMessage, status.ErrorCode)
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	// Vonage delivery receipt webhook.
	http.HandleFunc("/webhook/vonage", func(w http.ResponseWriter, r *http.Request) {
		status, err := vonage.ParseWebhook(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("[Vonage] Message %s → %s\n", status.MessageID, status.Status)

		if status.Status.IsFinal() {
			if status.Status.IsSuccess() {
				fmt.Printf("  Delivered\n")
			} else {
				fmt.Printf("  Failed: code=%s\n", status.ErrorCode)
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Webhook server listening on :8080")
	fmt.Println("  POST /webhook/twilio  — Twilio status callbacks")
	fmt.Println("  POST /webhook/vonage  — Vonage delivery receipts")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
