// Example: sending SMS via MSG91 using the Flow API.
//
// Set the following environment variables before running:
//
//	MSG91_AUTHKEY=your_authkey
//	MSG91_SENDER=SENDER          # 6-char DLT sender ID
//	MSG91_TEMPLATE_ID=tmpl_xxx   # DLT-approved Flow template ID
//	MSG91_TO=+919876543210
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
	"github.com/KARTIKrocks/gosms/msg91"
)

func main() {
	provider, err := msg91.NewProvider(msg91.Config{
		AuthKey:    os.Getenv("MSG91_AUTHKEY"),
		SenderID:   os.Getenv("MSG91_SENDER"),
		TemplateID: os.Getenv("MSG91_TEMPLATE_ID"),
		Route:      msg91.RouteTransactional,
	})
	if err != nil {
		log.Fatal(err)
	}

	client := gosms.NewClient(provider)
	ctx := context.Background()
	to := os.Getenv("MSG91_TO")

	// 1. Simple send — Body is passed as a default `body` variable for
	//    templates with a single ##body## placeholder.
	result, err := client.Send(ctx, to, "Welcome to gosms via MSG91!")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sent: request_id=%s status=%s\n", result.MessageID, result.Status)

	// 2. Template with variables. Your DLT template would reference
	//    ##name## and ##otp##; MSG91 substitutes them per recipient.
	msg := gosms.NewMessage(to, "")
	msg91.SetVar(msg, "name", "Kartik")
	msg91.SetVar(msg, "otp", "1234")

	result, err = client.SendMessage(ctx, msg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Templated: request_id=%s\n", result.MessageID)

	// 3. Bulk send — messages sharing a template ID go out in a single
	//    MSG91 API call with multiple recipients.
	batch := []*gosms.Message{
		msg91.SetVar(gosms.NewMessage("+919876500001", ""), "name", "Alice"),
		msg91.SetVar(gosms.NewMessage("+919876500002", ""), "name", "Bob"),
	}
	results, err := client.SendBulk(ctx, batch)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range results {
		fmt.Printf("Bulk: to=%s status=%s\n", r.To, r.Status)
	}

	// 4. Dedicated OTP flow via gosms.OTPProvider. MSG91 generates the
	//    code server-side when OTPRequest.OTP is empty.
	otpRes, err := provider.SendOTP(ctx, &gosms.OTPRequest{
		Phone:  to,
		Length: 6,
		Expiry: 5 * time.Minute,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("OTP sent: request_id=%s\n", otpRes.MessageID)

	// 5. Verify an OTP submitted by the user.
	if code := os.Getenv("MSG91_VERIFY_OTP"); code != "" {
		vr, err := provider.VerifyOTP(ctx, to, code)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("OTP verified=%v message=%q\n", vr.Verified, vr.Message)
	}

	// 6. Resend the last OTP over a different channel ("text" or "voice").
	if os.Getenv("MSG91_RESEND_OTP") != "" {
		if err := provider.ResendOTP(ctx, to, "voice"); err != nil {
			log.Fatal(err)
		}
		fmt.Println("OTP resent via voice")
	}
}
