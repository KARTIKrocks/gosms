---
id: vonage
title: Vonage
description: Configuring the Vonage provider and webhook parsing.
---

# Vonage

```bash
go get github.com/KARTIKrocks/gosms/vonage
```

## Configuration

```go
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

| Field | Purpose |
| --- | --- |
| `APIKey`, `APISecret` | Vonage API credentials |
| `From` | Default sender ID or number |
| `Type` | `vonage.TypeText` (default), `TypeUnicode`, or `TypeBinary` |
| `TTL` | Message time-to-live in milliseconds |
| `StatusReportRequired` | Request a delivery receipt callback |
| `CallbackURL` | Where Vonage posts delivery receipts |
| `HTTPClient` | Optional — override the default `*http.Client` |
| `BaseURL` | Optional — override the Vonage API base URL (testing) |

## Webhook

`vonage.ParseWebhook` matches the `gosms.WebhookParser` signature — see
[Webhooks](webhooks) for the shared pattern:

```go
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
