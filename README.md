<!-- The centred logo block opens the file, so there is no h1 on line 1. -->
<!-- markdownlint-disable-next-line MD041 -->
<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="website/static/img/logo-dark.svg">
    <img src="website/static/img/logo.svg" alt="gosms" width="96" height="96">
  </picture>
</p>

<h1 align="center">gosms</h1>

<p align="center">
  A unified SMS-sending library for Go, with support for Twilio, AWS SNS,
  Vonage, and MSG91 — each provider its own Go module.
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/KARTIKrocks/gosms"><img src="https://pkg.go.dev/badge/github.com/KARTIKrocks/gosms.svg" alt="Go Reference"></a>
  <a href="https://github.com/KARTIKrocks/gosms/releases"><img src="https://img.shields.io/github/v/tag/KARTIKrocks/gosms" alt="GitHub tag"></a>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/KARTIKrocks/gosms" alt="Go Version"></a>
  <a href="https://github.com/KARTIKrocks/gosms/actions/workflows/ci.yml"><img src="https://github.com/KARTIKrocks/gosms/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://codecov.io/gh/KARTIKrocks/gosms"><img src="https://codecov.io/gh/KARTIKrocks/gosms/branch/main/graph/badge.svg" alt="codecov"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License: MIT"></a>
</p>

<p align="center">
  <b><a href="https://kartikrocks.github.io/gosms/">Documentation</a></b> ·
  <b><a href="https://pkg.go.dev/github.com/KARTIKrocks/gosms">API Reference</a></b> ·
  <b><a href="CHANGELOG.md">Changelog</a></b>
</p>

## Why gosms?

Every provider ships its own SDK with its own request shape, its own error
strings, and its own webhook format. gosms gives every one of them the same
`Provider` interface, so an application sends a message the same way
regardless of which backend is behind it — and keeps each provider in its own
Go module, so switching (or supporting more than one) doesn't drag in
dependencies you don't use.

| Capability | gosms | Provider SDK directly |
| --- | --- | --- |
| One interface across Twilio, SNS, Vonage, and MSG91 | ✓ | You build it |
| Only pull the dependencies for providers you use | ✓ | N/A |
| Delivery status polling + webhook parsing | ✓ | You build it |
| Fallback and round-robin across providers | ✓ | You build it |
| E.164 validation, normalization, segment counting | ✓ | You build it |
| DLT template variables and OTP flow (MSG91) | ✓ | You build it |
| Mock provider with failure/latency injection | ✓ | You build it |

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

## Quick Start

```go
import (
    "github.com/KARTIKrocks/gosms"
    "github.com/KARTIKrocks/gosms/twilio"
)

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
if err != nil {
    log.Fatal(err)
}

log.Printf("Message sent: %s, Status: %s", result.MessageID, result.Status)
```

AWS SNS, Vonage, and MSG91 use the same `gosms.NewClient(provider)` /
`client.Send(...)` pattern once constructed — see the
[full guide for each provider](https://kartikrocks.github.io/gosms/docs/providers)
for their `Config` fields and specifics (MSG91 in particular is
template-driven rather than free-text).

## Features

- Unified `Provider` interface across all SMS backends
- Multi-module architecture — no unnecessary dependencies
- Message builder with fluent API
- Bulk messaging with `Batch` and `SendToMany`
- Delivery status tracking and webhook parsing
- Multi-provider with fallback and round-robin strategies
- Phone number validation (E.164) and normalization
- SMS segment calculation with proper GSM 03.38 charset support
- Pre-built message templates (OTP, alerts, notifications)
- Mock provider for testing

## Documentation

Full guides live at **[kartikrocks.github.io/gosms](https://kartikrocks.github.io/gosms/)**:

| Guide | Covers |
| --- | --- |
| [Getting Started](https://kartikrocks.github.io/gosms/docs/getting-started) | Install and send your first message |
| [Core](https://kartikrocks.github.io/gosms/docs/core) | `Provider` interface, `Client`, `Message` builder, `Result`/`Status` |
| [Providers](https://kartikrocks.github.io/gosms/docs/providers) | Twilio, AWS SNS, Vonage, and MSG91 configuration and specifics |
| [Multi-Provider](https://kartikrocks.github.io/gosms/docs/multi-provider) | Fallback and round-robin across providers |
| [Bulk Messaging](https://kartikrocks.github.io/gosms/docs/bulk) | `Batch`, `SendToMany`, and `Client.SendBulk` semantics |
| [Webhooks](https://kartikrocks.github.io/gosms/docs/webhooks) | Delivery-status callbacks for Twilio, Vonage, and MSG91 |
| [OTP](https://kartikrocks.github.io/gosms/docs/otp) | The optional `OTPProvider` send/verify/resend flow |
| [Helpers](https://kartikrocks.github.io/gosms/docs/helpers) | Phone validation/normalization, segment calculation, message templates |
| [Testing](https://kartikrocks.github.io/gosms/docs/testing) | `MockProvider` for tests |
| [Errors](https://kartikrocks.github.io/gosms/docs/errors) | Sentinel errors and the `DeliveryStatus` lifecycle |

Exact type signatures are always generated from source on
[pkg.go.dev](https://pkg.go.dev/github.com/KARTIKrocks/gosms).

## Examples

See the [`examples/`](examples/) directory for runnable examples:

| Example | Description |
| --- | --- |
| [basic](examples/basic/) | Core API usage with mock provider |
| [twilio-provider](examples/twilio-provider/) | Sending via Twilio |
| [sns-provider](examples/sns-provider/) | Sending via AWS SNS |
| [vonage-provider](examples/vonage-provider/) | Sending via Vonage |
| [msg91-provider](examples/msg91-provider/) | Sending via MSG91 (Flow templates + OTP) |
| [multi-provider](examples/multi-provider/) | Fallback and round-robin strategies |
| [webhooks](examples/webhooks/) | Delivery status webhook server |
| [mock-testing](examples/mock-testing/) | Using `MockProvider` in tests |
| [helpers](examples/helpers/) | Phone validation, normalization, segment calculation |

```bash
# Run an example (no credentials needed)
cd examples/basic && go run .
```

## Project Structure

```text
gosms/
├── sms.go          # Core types: Provider, Message, Result, Status, Client
├── helpers.go      # Utilities: validation, segments, batch, multi-provider
├── mock.go         # MockProvider for testing
├── doc.go          # Package documentation
├── twilio/         # Twilio provider (separate module)
├── sns/            # AWS SNS provider (separate module)
├── vonage/         # Vonage provider (separate module)
├── msg91/          # MSG91 provider (separate module)
├── website/        # Docusaurus documentation site (github.io/gosms)
└── examples/       # Runnable examples
```

## Thread Safety

- All providers are safe for concurrent use
- `Client` is safe for concurrent use after initialization
- `MockProvider` is safe for concurrent use with internal locking
- `MultiProvider` round-robin counter is atomic

## Security

Every push and pull request to `main` is scanned by
[CodeQL](https://github.com/KARTIKrocks/gosms/actions/workflows/codeql.yml),
with a full re-scan weekly to catch newly published query patterns against
unchanged code. [Dependabot](.github/dependabot.yml) tracks updates across
every Go module, the GitHub Actions themselves, and the documentation site's
npm dependencies.

See [SECURITY.md](SECURITY.md) for supported versions, what's in scope, and
how to report a vulnerability privately.

## License

[MIT](LICENSE)

## Contributing

Contributions welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
