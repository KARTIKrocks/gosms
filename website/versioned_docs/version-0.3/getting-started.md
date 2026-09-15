---
id: getting-started
title: Getting Started
description: Install gosms and send your first message.
---

# Getting Started

## Installation

Install the core module plus whichever providers you need:

```bash
# Core (required)
go get github.com/KARTIKrocks/gosms

# Providers (pick one or more)
go get github.com/KARTIKrocks/gosms/twilio
go get github.com/KARTIKrocks/gosms/sns
go get github.com/KARTIKrocks/gosms/vonage
go get github.com/KARTIKrocks/gosms/msg91
```

Each provider is its own Go module, so `go get` only pulls the dependencies
that provider actually needs — a Twilio-only app never sees the AWS SDK.

## Send your first message

```go
package main

import (
    "context"
    "log"

    "github.com/KARTIKrocks/gosms"
    "github.com/KARTIKrocks/gosms/twilio"
)

func main() {
    provider, err := twilio.NewProvider(twilio.Config{
        AccountSID: "account_sid",
        AuthToken:  "auth_token",
        From:       "+15551234567",
    })
    if err != nil {
        log.Fatal(err)
    }

    client := gosms.NewClient(provider)

    result, err := client.Send(context.Background(), "+15559876543", "Hello from gosms!")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Message sent: %s, Status: %s", result.MessageID, result.Status)
}
```

`gosms.NewClient` wraps any `Provider` — see [Core](core) for the full
`Client`/`Message`/`Result` contract, and [Providers](providers) for the other
three backends.

## Testing without credentials

Every example in this guide can be exercised without a provider account using
the built-in [`MockProvider`](testing):

```go
mock := gosms.NewMockProvider()
client := gosms.NewClient(mock)

result, err := client.Send(ctx, "+15551234567", "Test message")
```

## Runnable examples

The [`examples/`](https://github.com/KARTIKrocks/gosms/tree/main/examples)
directory has a runnable program for every provider plus multi-provider,
webhooks, bulk sending, and helper usage:

```bash
git clone https://github.com/KARTIKrocks/gosms.git
cd gosms/examples/basic
go run .
```
