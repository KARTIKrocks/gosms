---
id: msg91
title: MSG91
description: DLT Flow templates, template variables, phone normalization, and bulk chunking for MSG91.
---

# MSG91

MSG91 is the standard SMS gateway for India. Unlike the other providers, it
is **template-driven** — it sends DLT-approved Flow templates, not free text.

```bash
go get github.com/KARTIKrocks/gosms/msg91
```

## Configuration

```go
provider, err := msg91.NewProvider(msg91.Config{
    AuthKey:    "your_authkey",
    SenderID:   "SENDER",         // 6-char DLT sender ID
    TemplateID: "tmpl_xxx",       // DLT-approved Flow template
    Route:      msg91.RouteTransactional,
})
if err != nil {
    log.Fatal(err)
}

client := gosms.NewClient(provider)

msg := gosms.NewMessage("+919876543210", "")
msg91.SetVar(msg, "name", "Kartik")
msg91.SetVar(msg, "otp", "1234")

result, err := client.SendMessage(ctx, msg)
```

| Field | Purpose |
| --- | --- |
| `AuthKey` | MSG91 API auth key |
| `SenderID` | 6-character DLT-registered sender ID |
| `TemplateID` | Default DLT-approved Flow template |
| `Route` | `msg91.RouteTransactional` (default, route 4) or `msg91.RoutePromotional` (route 1) |
| `Country` | Default country code used to normalize short numbers — see below (default `91`) |
| `ShortURL` | Enable MSG91's URL shortening |
| `MaxRecipientsPerCall` | Chunk size for `SendBulk` (default `1000`; negative disables chunking) |
| `HTTPClient` | Optional — override the default `*http.Client` |
| `BaseURL` | Optional — override the MSG91 API base URL (testing) |

## Template variables and `Body` fallback

MSG91 Flow templates reference placeholders like `##name##` or `##otp##`. Set
each one with `msg91.SetVar`. For templates with a single `##body##`
placeholder, any non-empty `Message.Body` is automatically passed as `body`
when no vars are set — so the unified `client.Send(ctx, to, text)` path works
without extra wiring.

## Per-message overrides

```go
msg91.SetTemplateID(msg, "tmpl_other")  // override Config.TemplateID
msg.WithFrom("OTHER")                   // override the sender ID
```

## Phone normalization

Non-E.164 numbers up to 10 digits are prefixed with `Config.Country` (default
`91`). `+919876543210`, `919876543210`, and `9876543210` all normalize
identically.

## Bulk chunking

`SendBulk` groups recipients by effective `(template_id, sender)` and sends
each group as one Flow API call. Groups larger than `Config.MaxRecipientsPerCall`
(default 1000) are automatically chunked across multiple calls; set a
negative value to disable chunking.

## OTP

MSG91 implements the optional [`gosms.OTPProvider`](otp) capability for the
full send / verify / resend flow, including server-side code generation. See
[OTP](otp) for the full example.

## Webhook

Delivery status is webhook-driven; parse incoming callbacks with
`msg91.ParseWebhook(r)` — see [Webhooks](webhooks) for the shared pattern.
