# gosms

[![Go Reference](https://pkg.go.dev/badge/github.com/KARTIKrocks/gosms.svg)](https://pkg.go.dev/github.com/KARTIKrocks/gosms)
[![Go Report Card](https://goreportcard.com/badge/github.com/KARTIKrocks/gosms)](https://goreportcard.com/report/github.com/KARTIKrocks/gosms)
[![Go Version](https://img.shields.io/github/go-mod/go-version/KARTIKrocks/gosms)](go.mod)
[![CI](https://github.com/KARTIKrocks/gosms/actions/workflows/ci.yml/badge.svg)](https://github.com/KARTIKrocks/gosms/actions/workflows/ci.yml)
[![GitHub tag](https://img.shields.io/github/v/tag/KARTIKrocks/gosms)](https://github.com/KARTIKrocks/gosms/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![codecov](https://codecov.io/gh/KARTIKrocks/gosms/branch/main/graph/badge.svg)](https://codecov.io/gh/KARTIKrocks/gosms)

A unified SMS sending library for Go with support for multiple providers including Twilio, AWS SNS, and Vonage.

Each provider is a **separate Go module**, so you only download the dependencies you actually need.

## Installation

Install only what you need:

```bash
# Core (required)
go get github.com/KARTIKrocks/gosms

# Providers (pick one or more)
go get github.com/KARTIKrocks/gosms/twilio
go get github.com/KARTIKrocks/gosms/sns
go get github.com/KARTIKrocks/gosms/vonage
```

## Quick Start

### Twilio

```go
import (
    "github.com/KARTIKrocks/gosms"
    "github.com/KARTIKrocks/gosms/twilio"
)

provider, err := twilio.NewProvider(twilio.Config{
    AccountSID: "account_sid",
    AuthToken:  "auth_token",
    From:       "+15551234567",
})
if err != nil {
    log.Fatal(err)
}

client := gosms.NewClient(provider)

result, err := client.Send(ctx, "+15559876543", "Hello from gosms!")
if err != nil {
    log.Fatal(err)
}

log.Printf("Message sent: %s, Status: %s", result.MessageID, result.Status)
```

### AWS SNS

```go
import (
    "github.com/KARTIKrocks/gosms"
    "github.com/KARTIKrocks/gosms/sns"
)

config := sns.DefaultConfig()
config.Region = "us-east-1"
config.AccessKeyID = "access_key"
config.SecretAccessKey = "secret_key"
config.SenderID = "MyApp"
config.SMSType = sns.SMSTransactional

provider, err := sns.NewProvider(ctx, config)
if err != nil {
    log.Fatal(err)
}

client := gosms.NewClient(provider)
result, err := client.Send(ctx, "+15559876543", "Your code is 123456")
```

### Vonage

```go
import (
    "github.com/KARTIKrocks/gosms"
    "github.com/KARTIKrocks/gosms/vonage"
)

provider, err := vonage.NewProvider(vonage.Config{
    APIKey:    "api_key",
    APISecret: "api_secret",
    From:      "MyApp",
})
if err != nil {
    log.Fatal(err)
}

client := gosms.NewClient(provider)
result, err := client.Send(ctx, "+15559876543", "Hello from Vonage!")
```

## Features

- Unified `Provider` interface across all SMS backends
- Multi-module architecture — no unnecessary dependencies
- Message builder with fluent API
- Bulk messaging with `Batch` and `SendToMany`
- Delivery status tracking and webhook parsing
- Multi-provider with fallback and round-robin strategies
- Phone number validation (E.164) and normalization
- SMS segment calculation with proper GSM 03.38 charset support
- Pre-built message templates (OTP, alerts, notifications)
- Mock provider for testing

## Message Builder

```go
msg := gosms.NewMessage("+15559876543", "Hello!").
    WithFrom("+15551234567").
    WithReference("order-123").
    WithValidity(1 * time.Hour).
    WithMetadata("user_id", "12345")

result, err := client.SendMessage(ctx, msg)
```

### Scheduled Messages (Twilio)

```go
msg := gosms.NewMessage("+15559876543", "Reminder: Your appointment is tomorrow").
    WithSchedule(time.Now().Add(24 * time.Hour))

result, err := client.SendMessage(ctx, msg)
```

## Bulk Messaging

### Using Batch

```go
batch := gosms.NewBatch()
batch.AddNew("+15551111111", "Message 1")
batch.AddNew("+15552222222", "Message 2")
batch.AddNew("+15553333333", "Message 3")

results, err := batch.Send(ctx, client)
for _, result := range results {
    if result.Success() {
        log.Printf("Sent to %s: %s", result.To, result.MessageID)
    } else {
        log.Printf("Failed to %s: %s", result.To, result.Error)
    }
}
```

### Send Same Message to Many

```go
results, err := gosms.SendToMany(ctx, client,
    "Flash sale! 50% off today only!",
    "+15551111111",
    "+15552222222",
    "+15553333333",
)
```

## Delivery Status

### Get Status

```go
status, err := client.GetStatus(ctx, "message_id")
if err != nil {
    log.Fatal(err)
}

if status.Status.IsFinal() {
    if status.Status.IsSuccess() {
        log.Printf("Message delivered at %v", status.UpdatedAt)
    } else {
        log.Printf("Delivery failed: %s", status.ErrorMessage)
    }
}
```

### Twilio Webhook

```go
import "github.com/KARTIKrocks/gosms/twilio"

http.HandleFunc("/webhook/twilio", func(w http.ResponseWriter, r *http.Request) {
    status, err := twilio.ParseWebhook(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    log.Printf("Message %s status: %s", status.MessageID, status.Status)
    w.WriteHeader(http.StatusOK)
})
```

### Vonage Webhook

```go
import "github.com/KARTIKrocks/gosms/vonage"

http.HandleFunc("/webhook/vonage", func(w http.ResponseWriter, r *http.Request) {
    status, err := vonage.ParseWebhook(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    log.Printf("Message %s status: %s", status.MessageID, status.Status)
    w.WriteHeader(http.StatusOK)
})
```

## Multi-Provider

### Fallback

Try Twilio first, fall back to Vonage on failure:

```go
multi := gosms.NewMultiProvider(twilioProvider, vonageProvider)

client := gosms.NewClient(multi)
result, err := client.Send(ctx, to, body)
```

### Round-Robin

Rotate across providers:

```go
multi := gosms.NewMultiProvider(twilioProvider, vonageProvider).
    WithStrategy(gosms.StrategyRoundRobin)

client := gosms.NewClient(multi)
```

## Phone Number Utilities

```go
// Validate E.164 format
if gosms.ValidateE164("+15551234567") {
    log.Println("Valid E.164 number")
}

// Normalize phone number
normalized := gosms.NormalizePhone("555-123-4567", "+1")
// Returns: +15551234567

// Check if message uses GSM 7-bit encoding
if gosms.IsGSMEncoding("Hello world") {
    log.Println("GSM encoding (160 char limit)")
}

// Calculate SMS segments
segments := gosms.CalculateSegments("Hello, this is a test message!")
log.Printf("Message will use %d segment(s)", segments)
```

## Pre-built Message Templates

```go
// OTP: "123456 is your MyApp verification code."
msg := gosms.OTPMessage("+15551234567", "123456", "MyApp")

// Alert: "[URGENT] Server is down!"
msg := gosms.AlertMessage("+15551234567", "URGENT", "Server is down!")

// Notification: "Order Update: Your order has shipped"
msg := gosms.NotificationMessage("+15551234567", "Order Update", "Your order has shipped")
```

## Testing with Mock Provider

```go
mock := gosms.NewMockProvider()
client := gosms.NewClient(mock)

// Send message
result, err := client.Send(ctx, "+15551234567", "Test message")

// Verify
if mock.MessageCount() != 1 {
    t.Error("Expected 1 message")
}

lastMsg := mock.LastMessage()
if lastMsg.Message.Body != "Test message" {
    t.Error("Message body mismatch")
}

// Simulate failures
mock.WithFailAll(true)
result, err = client.Send(ctx, "+15551234567", "This will fail")
// result.Status == gosms.StatusFailed

// Simulate errors
mock.WithSendError(gosms.ErrRateLimited)
_, err = client.Send(ctx, "+15551234567", "This will error")
// errors.Is(err, gosms.ErrRateLimited) == true

// Reset mock
mock.Reset()
```

## Error Handling

```go
result, err := client.Send(ctx, to, body)
if err != nil {
    switch {
    case errors.Is(err, gosms.ErrInvalidPhone):
        log.Println("Invalid phone number")
    case errors.Is(err, gosms.ErrInvalidMessage):
        log.Println("Invalid message content")
    case errors.Is(err, gosms.ErrRateLimited):
        log.Println("Rate limited, try again later")
    case errors.Is(err, gosms.ErrInsufficientFunds):
        log.Println("Account balance too low")
    case errors.Is(err, gosms.ErrBlacklisted):
        log.Println("Number is blacklisted")
    case errors.Is(err, gosms.ErrProviderError):
        log.Println("Provider error:", err)
    default:
        log.Println("Unknown error:", err)
    }
}
```

## Delivery Status Values

| Status            | Description                     |
| ----------------- | ------------------------------- |
| `StatusPending`   | Message is pending              |
| `StatusQueued`    | Message is queued for delivery  |
| `StatusAccepted`  | Message accepted by provider    |
| `StatusSent`      | Message sent to carrier         |
| `StatusDelivered` | Message delivered to recipient  |
| `StatusFailed`    | Delivery failed                 |
| `StatusRejected`  | Message was rejected            |
| `StatusExpired`   | Message expired before delivery |
| `StatusUnknown`   | Status unknown                  |

## AWS SNS Additional Features

```go
import "github.com/KARTIKrocks/gosms/sns"

provider, _ := sns.NewProvider(ctx, config)

// Set account-level SMS attributes
err := provider.SetSMSAttributes(ctx,
    "100.00",                          // Monthly spend limit
    "arn:aws:iam::123:role/SNSRole",   // IAM role for delivery logs
    "100",                             // Success sampling rate %
)

// Check opt-out status
optedOut, err := provider.CheckIfPhoneNumberIsOptedOut(ctx, "+15551234567")

// List opted-out numbers
numbers, err := provider.ListPhoneNumbersOptedOut(ctx)

// Opt-in a number
err = provider.OptInPhoneNumber(ctx, "+15551234567")
```

## Examples

See the [`examples/`](examples/) directory for runnable examples:

| Example                                      | Description                                          |
| -------------------------------------------- | ---------------------------------------------------- |
| [basic](examples/basic/)                     | Core API usage with mock provider                    |
| [twilio-provider](examples/twilio-provider/) | Sending via Twilio                                   |
| [sns-provider](examples/sns-provider/)       | Sending via AWS SNS                                  |
| [vonage-provider](examples/vonage-provider/) | Sending via Vonage                                   |
| [multi-provider](examples/multi-provider/)   | Fallback and round-robin strategies                  |
| [webhooks](examples/webhooks/)               | Delivery status webhook server                       |
| [mock-testing](examples/mock-testing/)       | Using MockProvider in tests                          |
| [helpers](examples/helpers/)                 | Phone validation, normalization, segment calculation |

```bash
# Run an example (no credentials needed)
cd examples/basic && go run .
```

## Project Structure

```
gosms/
├── sms.go          # Core types: Provider, Message, Result, Status, Client
├── helpers.go      # Utilities: validation, segments, batch, multi-provider
├── mock.go         # MockProvider for testing
├── doc.go          # Package documentation
├── twilio/         # Twilio provider (separate module)
│   ├── go.mod
│   └── twilio.go
├── sns/            # AWS SNS provider (separate module)
│   ├── go.mod
│   └── sns.go
├── vonage/         # Vonage provider (separate module)
│   ├── go.mod
│   └── vonage.go
└── examples/       # Runnable examples
    ├── basic/
    ├── twilio-provider/
    ├── sns-provider/
    ├── vonage-provider/
    ├── multi-provider/
    ├── webhooks/
    ├── mock-testing/
    └── helpers/
```

## Thread Safety

- All providers are safe for concurrent use
- `Client` is safe for concurrent use after initialization
- `MockProvider` is safe for concurrent use with internal locking
- `MultiProvider` round-robin counter is atomic

## License

MIT
