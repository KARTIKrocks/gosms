---
id: testing
title: Testing with MockProvider
description: Using the in-memory MockProvider for tests, with failure and latency injection.
---

# Testing with MockProvider

`MockProvider` is an in-memory `Provider` in the core module — no network
calls, no credentials, safe for concurrent use.

```go
mock := gosms.NewMockProvider()
client := gosms.NewClient(mock)

result, err := client.Send(ctx, "+15551234567", "Test message")

if mock.MessageCount() != 1 {
    t.Error("Expected 1 message")
}

lastMsg := mock.LastMessage()
if lastMsg.Message.Body != "Test message" {
    t.Error("Message body mismatch")
}
```

## Injecting failures

```go
// Every Send returns StatusFailed.
mock.WithFailAll(true)
result, err = client.Send(ctx, "+15551234567", "This will fail")

// Every Send returns this error instead of succeeding.
mock.WithSendError(gosms.ErrRateLimited)
_, err = client.Send(ctx, "+15551234567", "This will error")
// errors.Is(err, gosms.ErrRateLimited) == true

// GetStatus returns this error instead of a status.
mock.WithStatusError(gosms.ErrProviderError)

// Simulate network latency (respects context cancellation).
mock.WithLatency(50 * time.Millisecond)

// Whether accepted messages should also resolve to StatusDelivered on
// GetStatus (default true).
mock.WithDeliverAll(false)
```

## Inspecting sent messages

| Method | Returns |
| --- | --- |
| `Messages()` | All sent `*MockMessage`s, in order |
| `LastMessage()` | The most recent `*MockMessage`, or `nil` if none sent |
| `MessageCount()` | Number of messages sent |
| `FindMessagesByTo(to)` | All messages sent to a given recipient |
| `FindMessageByID(id)` | The message with the given provider message ID, or `nil` |
| `SetStatus(id, status)` | Manually set the status returned by `GetStatus` for a message ID |

`MockMessage` carries `MessageID`, `Message` (the original `*gosms.Message`),
`SentAt`, and `Status`.

## Resetting between tests

```go
mock.Clear()  // clears sent messages and statuses only
mock.Reset()  // clears everything, including injected failures/latency
```

Prefer `Reset()` in `t.Cleanup` when a test configures failure injection, so
state never leaks into the next test.
