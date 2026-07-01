# AGENTS.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`gosms` is a unified SMS-sending library for Go supporting multiple providers (Twilio, AWS SNS, Vonage, MSG91). Each provider lives in its **own Go module** so consumers only pull the dependencies for the providers they use.

## Multi-module layout

This is a **Go workspace** (`go.work`) of five modules:

- `.` (root) — `github.com/KARTIKrocks/gosms`: core types and provider-independent helpers. **No third-party dependencies.**
- `./twilio`, `./sns`, `./vonage`, `./msg91` — one module per provider, each importing the core module.

Implications:

- Every module has its own `go.mod`/`go.sum`. Run `go` commands from inside the relevant module directory; the `Makefile` loops over all modules for you.
- The `go.work` file makes provider modules resolve the core module locally during development, so changes to core are immediately visible to providers without publishing.
- Releases are versioned per-module via git tags (`v0.1.0`, `twilio/v0.1.0`, etc.). See the release workflow below.

## Common commands

All run from the repo root unless noted. The `Makefile` is the canonical entry point and iterates over every module.

```bash
make ci          # what CI runs: fmt-check, vet, lint, test-race
make test        # go test ./... in every module
make test-race   # tests with -race -count=1
make lint        # golangci-lint across all modules (auto-installs tools via `make setup`)
make fmt         # gofmt -s + goimports (writes); fmt-check is the non-mutating CI variant
make build       # go build ./... in every module
make coverage    # merged coverage across modules -> coverage.out
make bench       # benchmarks across all modules
```

Targeting a single module or test (run inside the module dir):

```bash
cd twilio && go test ./...                  # one module
cd twilio && go test -run TestSend ./...    # one test by name
go test -run TestCalculateSegments ./...    # a core test (from root)
```

`make setup` installs pinned `golangci-lint` (v2) and `goimports` versions; `make lint`/`make fmt` depend on it.

## Architecture

The core module (`sms.go`, `helpers.go`, `mock.go`) defines the contract; provider modules implement it.

- **`Provider` interface** (`sms.go`) — the integration seam every provider implements: `Send`, `SendBulk`, `GetStatus`, `Name`. A provider package's `NewProvider(...)` returns a value satisfying this interface.
- **`Client`** (`sms.go`) — thin wrapper over a single `Provider`. `Send(ctx, to, body)` is the convenience path; `SendMessage(ctx, *Message)` is the full path. The client validates messages and applies `defaultFrom` before delegating. `SendBulk` validates per-message and records validation failures as failed `Result`s rather than aborting the batch.
- **Optional capabilities via interface assertion** — beyond the base `Provider`, the core defines opt-in interfaces a provider _may_ also implement, detected with a type assertion:
  - `OTPProvider` (`SendOTP`/`VerifyOTP`/`ResendOTP`) — currently implemented by MSG91.
  - `WebhookParser` — a core func type (`sms.go`) documenting the conventional signature `func(*http.Request) (*Status, error)`; it is not an interface on Provider. The Twilio, Vonage, and MSG91 packages each expose a package-level `ParseWebhook` matching it. SNS does not.
- **`MultiProvider`** (`helpers.go`) — itself a `Provider` that wraps several providers with `StrategyFallback` (try in order) or `StrategyRoundRobin` (atomic counter). Compose providers by passing a `MultiProvider` to `NewClient`.
- **`MockProvider`** (`mock.go`) — in-memory `Provider` for tests; supports failure/error injection (`WithFailAll`, `WithSendError`) and message inspection (`LastMessage`, `Reset`).
- **Provider-agnostic helpers** (`helpers.go`) — E.164 validation/normalization, GSM 03.38 vs UCS-2 segment counting (`CalculateSegments`), `Batch` builder, and message constructors (`OTPMessage`, `AlertMessage`, etc.). These have no provider dependencies.

### Data flow

`Client.Send` → `NewMessage` → `Message.Validate` → `Provider.Send` → `*Result`. Statuses use the `DeliveryStatus` enum; `Result.Success()` means _accepted/sent/delivered_, while `DeliveryStatus.IsSuccess()` means _confirmed delivered only_. `IsFinal()` marks terminal states.

### Provider-specific notes

- **MSG91** is template-driven (DLT Flow templates), not free-text. Variables are set with `msg91.SetVar(msg, key, val)` rather than `Message.Body`; a single-placeholder template falls back to `Body`. Per-message overrides via `msg91.SetTemplateID`. It also implements `OTPProvider` (server-side code generation when `OTPRequest.OTP` is empty).
- **SNS** uses `sns.DefaultConfig()` + `NewProvider(ctx, config)` (note the `ctx` arg, unlike other providers) and exposes account-level extras (`SetSMSAttributes`, opt-out management).

## Conventions

- Errors are sentinel values in `sms.go` (`ErrInvalidPhone`, `ErrRateLimited`, etc.); providers must wrap to these so callers can use `errors.Is`. Prefer extending this set over inventing new error styles.
- All exported types/functions require doc comments (enforced by lint).
- All providers and `Client`/`MockProvider`/`MultiProvider` are documented as safe for concurrent use — preserve that when modifying.
- A change to a core type usually requires touching all provider modules; run `make ci` (which covers every module) before considering a change complete.

## Releasing (per-module tags)

`make release-prep VERSION=vX.Y.Z` strips the local `replace` directives and pins sub-modules to the tagged core version; `make release-local` restores `replace` directives for local development. The prep target prints the exact `git tag` commands (one tag per module).
