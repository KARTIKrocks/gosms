---
id: twilio
title: Twilio
description: Configuring the Twilio provider, scheduled messages, and webhook parsing.
---

# Twilio

```bash
go get github.com/KARTIKrocks/gosms/twilio
```

## Configuration

```go
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
```

| Field | Purpose |
| --- | --- |
| `AccountSID`, `AuthToken` | Twilio API credentials |
| `From` | Default sender number, used when a `Message` doesn't set `From` |
| `MessagingServiceSID` | Optional — route through a Messaging Service instead of a single `From` number |
| `StatusCallback` | Optional — URL Twilio posts delivery status updates to |
| `HTTPClient` | Optional — override the default `*http.Client` |
| `BaseURL` | Optional — override the Twilio API base URL (testing) |

## Scheduled messages

Twilio is the only provider that supports `WithSchedule`:

```go
msg := gosms.NewMessage("+15559876543", "Reminder: your appointment is tomorrow").
    WithSchedule(time.Now().Add(24 * time.Hour))

result, err := client.SendMessage(ctx, msg)
```

## Webhook

`twilio.ParseWebhook` matches the `gosms.WebhookParser` signature — see
[Webhooks](webhooks) for the shared pattern:

```go
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
