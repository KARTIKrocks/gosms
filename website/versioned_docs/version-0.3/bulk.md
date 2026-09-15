---
id: bulk
title: Bulk Messaging
description: Sending to many recipients with Batch, SendToMany, and Client.SendBulk.
---

# Bulk Messaging

## `Client.SendBulk`

`SendBulk` validates each message independently. A message that fails
validation is recorded as a failed `Result` at its original index rather than
aborting the batch — the rest still send:

```go
results, err := client.SendBulk(ctx, msgs)
```

## Using `Batch`

`Batch` is a small builder around a slice of `*Message`:

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

`Batch` also has `Add(msg)`, `AddNewWithFrom(to, body, from)`, `Messages()`,
`Size()`, and `Clear()`.

## Send the same message to many recipients

```go
results, err := gosms.SendToMany(ctx, client,
    "Flash sale! 50% off today only!",
    "+15551111111",
    "+15552222222",
    "+15553333333",
)
```

## Provider-level chunking

Some providers chunk large batches internally to match a provider API limit —
see [MSG91's bulk chunking](msg91#bulk-chunking) for the worked example. This
is transparent to `Client.SendBulk`: one `*Result` per input `*Message` comes
back either way.
