---
id: core
title: Core
description: The Provider interface, Client wrapper, Message builder, and Result/Status types.
---

# Core

Everything on this page lives in the root module (`github.com/KARTIKrocks/gosms`)
and has no third-party dependencies.

## Provider interface

Every backend — Twilio, SNS, Vonage, MSG91, and the [mock provider](testing) —
implements the same interface. Anything that satisfies it is a drop-in
provider:

```go
type Provider interface {
    Send(ctx context.Context, msg *Message) (*Result, error)
    SendBulk(ctx context.Context, msgs []*Message) ([]*Result, error)
    GetStatus(ctx context.Context, messageID string) (*Status, error)
    Name() string
}
```

Beyond this base interface, a provider may also implement one of two optional
capabilities, detected with a type assertion:

- [`OTPProvider`](otp) — `SendOTP` / `VerifyOTP` / `ResendOTP`. Currently
  implemented by MSG91.
- `WebhookParser` — not an interface on `Provider`, but a documented func
  signature (`func(*http.Request) (*Status, error)`) that Twilio, Vonage, and
  MSG91 each expose as a package-level `ParseWebhook`. See [Webhooks](webhooks).

## Client

`Client` wraps any `Provider` and adds convenience methods. It is safe for
concurrent use after construction.

| Method | Description |
| --- | --- |
| `Send(ctx, to, body)` | Send plain text — shortcut over `SendMessage` |
| `SendMessage(ctx, msg)` | Send a fully-built `Message` (overrides, schedule, metadata) |
| `SendBulk(ctx, msgs)` | Send a slice of `Message`s — see [Bulk Messaging](bulk) |
| `GetStatus(ctx, id)` | Look up delivery status by provider message ID |
| `Provider()` | Access the underlying `Provider` (for type assertions like `OTPProvider`) |
| `ProviderName()` | The underlying provider's `Name()` |
| `WithDefaultFrom(from)` | Set a default sender applied when a `Message` doesn't set one |

```go
client := gosms.NewClient(provider).WithDefaultFrom("+15551234567")

result, err := client.Send(ctx, "+15559876543", "Hello!")
```

`Client.Send` and `SendMessage` validate the message (`Message.Validate`)
before delegating to the provider; an invalid message never reaches the
network. `SendBulk` validates per-message and records validation failures as
failed `Result`s rather than aborting the whole batch — see [Bulk Messaging](bulk).

## Message builder

Build a message with the fluent API. Every `With*` method returns `*Message`
for chaining.

```go
msg := gosms.NewMessage("+15559876543", "Hello!").
    WithFrom("+15551234567").
    WithReference("order-123").
    WithValidity(1 * time.Hour).
    WithMetadata("user_id", "12345")

result, err := client.SendMessage(ctx, msg)
```

| Field / method | Purpose |
| --- | --- |
| `To`, `Body` | Set via `NewMessage(to, body)` |
| `WithFrom(from)` | Sender ID or phone number |
| `WithReference(ref)` | Client-supplied reference, echoed back where the provider supports it |
| `WithSchedule(t)` | Schedule for future delivery (Twilio) |
| `WithValidity(d)` | How long the message stays valid for delivery |
| `WithMetadata(key, val)` | Arbitrary provider-specific data |

## Result and Status

Every send returns a `*Result`:

```go
type Result struct {
    MessageID string         // Provider-assigned ID
    To        string         // Recipient
    Status    DeliveryStatus
    Provider  string
    Cost      string
    Currency  string
    Segments  int
    SentAt    time.Time
    Error     string         // Set on per-message failure (bulk)
    Raw       map[string]any // Raw provider response
}

if result.Success() {
    log.Printf("Sent to %s: %s", result.To, result.MessageID)
}
```

`Result.Success()` means the provider **accepted** the message (accepted,
sent, or delivered) — not that it was confirmed delivered. For confirmed
delivery, check `DeliveryStatus.IsSuccess()` on a `Status` fetched via
`GetStatus` or a webhook. See [Errors](errors) for the full `DeliveryStatus`
lifecycle.
