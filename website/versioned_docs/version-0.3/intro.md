---
id: intro
title: Introduction
description: What gosms is, why it exists, and how the modules fit together.
---

# gosms

**gosms** is a unified SMS-sending library for Go. It gives every supported
provider — [Twilio](https://www.twilio.com/), [AWS SNS](https://aws.amazon.com/sns/),
[Vonage](https://www.vonage.com/), and [MSG91](https://msg91.com/) — the same
`Provider` interface, so application code sends a message the same way
regardless of which backend is behind it.

## Why gosms?

Every provider ships its own SDK with its own request shape, its own error
strings, and its own webhook payload format. Writing directly against one
means:

- Rewriting the integration if you ever add a second provider, or fall back
  to one when another is down.
- Re-deriving phone validation, SMS segment counting, and delivery-status
  polling for every provider you add.
- Pulling in every provider's SDK dependencies even when you only use one.

gosms solves this with one `Provider` interface (`Send`, `SendBulk`,
`GetStatus`, `Name`), a shared sentinel error set matched with `errors.Is`,
provider-agnostic helpers (E.164 validation, segment counting, message
builders), and a `MultiProvider` that composes several providers behind the
same interface with fallback or round-robin routing.

## Multi-module layout

gosms is a Go workspace of five modules:

- `github.com/KARTIKrocks/gosms` (the core) — the `Provider` interface,
  `Client`, `Message`, `Result`, sentinel errors, and provider-agnostic
  helpers. **No third-party dependencies.**
- `github.com/KARTIKrocks/gosms/twilio`, `/sns`, `/vonage`, `/msg91` — one
  module per provider, each importing only the core module.

Each provider is a separate `go get`, so an application using only Twilio
never pulls in the AWS SDK, and vice versa.

## Where to go next

- [Getting Started](getting-started) — install and send your first message.
- [Core](core) — the `Provider` interface, `Client`, and `Message` builder.
- [Providers](providers) — per-provider configuration and specifics.
- [Errors](errors) — the sentinel error set and delivery-status lifecycle.

Exact type signatures are always generated from source on
[pkg.go.dev](https://pkg.go.dev/github.com/KARTIKrocks/gosms). These guides
cover concepts and worked examples, and link out for the exact signatures.
