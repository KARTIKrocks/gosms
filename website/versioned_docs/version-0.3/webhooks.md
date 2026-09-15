---
id: webhooks
title: Webhooks
description: Parsing provider delivery-status callbacks with the shared WebhookParser convention.
---

# Webhooks

Three of the four providers deliver status updates via webhook rather than
(or in addition to) `GetStatus` polling: Twilio, Vonage, and MSG91. Each
exposes a package-level `ParseWebhook` function matching the core
`gosms.WebhookParser` signature:

```go
type WebhookParser func(r *http.Request) (*Status, error)
```

This lets callers store or route parsers generically:

```go
parsers := map[string]gosms.WebhookParser{
    "/webhook/twilio": twilio.ParseWebhook,
    "/webhook/vonage": vonage.ParseWebhook,
    "/webhook/msg91":  msg91.ParseWebhook,
}
```

## Twilio

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

## Vonage

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

## MSG91

Delivery status for MSG91 is webhook-driven — there is no status-polling
alternative. Use `msg91.ParseWebhook(r)` the same way as above.

## AWS SNS

SNS does not implement `WebhookParser`. Its `GetStatus` returns
`gosms.ErrUnsupported`; use CloudWatch delivery-status logging instead — see
[AWS SNS → Delivery status](sns#delivery-status).
