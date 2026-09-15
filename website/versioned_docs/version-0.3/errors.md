---
id: errors
title: Errors
description: The sentinel error set and the DeliveryStatus lifecycle.
---

# Errors

## Sentinel errors

All defined in the root module (`sms.go`). Providers wrap their own errors to
these so callers can match with `errors.Is`:

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

| Error | Meaning |
| --- | --- |
| `ErrInvalidConfig` | Provider configuration failed validation |
| `ErrInvalidPhone` | Recipient (or sender) phone number is invalid |
| `ErrInvalidMessage` | Message content failed validation (e.g. empty body) |
| `ErrSendFailed` | Generic send failure not covered by a more specific error |
| `ErrProviderError` | The provider's API returned an error |
| `ErrRateLimited` | The provider is rate-limiting this account |
| `ErrInsufficientFunds` | Account balance too low to send |
| `ErrBlacklisted` | Recipient number is blacklisted |
| `ErrUnsupported` | The provider doesn't support the requested operation (e.g. SNS's `GetStatus`) |

## Delivery status

`DeliveryStatus` is a string enum describing where a message is in its
lifecycle:

| Status | Description |
| --- | --- |
| `StatusPending` | Message is pending |
| `StatusQueued` | Message is queued for delivery |
| `StatusAccepted` | Message accepted by provider |
| `StatusSent` | Message sent to carrier |
| `StatusDelivered` | Message delivered to recipient |
| `StatusFailed` | Delivery failed |
| `StatusRejected` | Message was rejected |
| `StatusExpired` | Message expired before delivery |
| `StatusUnknown` | Status unknown |

Two helpers classify a status:

```go
status, err := client.GetStatus(ctx, "message_id")
if status.Status.IsFinal() {
    if status.Status.IsSuccess() {
        log.Printf("Message delivered at %v", status.UpdatedAt)
    } else {
        log.Printf("Delivery failed: %s", status.ErrorMessage)
    }
}
```

- **`IsFinal()`** — true for `StatusDelivered`, `StatusFailed`,
  `StatusRejected`, `StatusExpired`. No further status updates are expected.
- **`IsSuccess()`** — true only for `StatusDelivered` (confirmed delivery).

Don't confuse this with `Result.Success()` (see [Core](core#result-and-status)),
which means the provider **accepted** the message — not that it was
confirmed delivered.
