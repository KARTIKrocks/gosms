# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-09-08

### Changed

- **Minimum Go version is now 1.27** (was 1.24), across the core module, every provider module, and the examples. Consumers must build with Go 1.27 or newer.
- Time-dependent `MockProvider` latency tests now use `testing/synctest`, so they run deterministically with a fake clock instead of sleeping in real time.
- Expanded `.golangci.yml` with security (`gosec`), error-handling (`errorlint`, `errname`, `nilerr`, `nilnesserr`), context/concurrency (`contextcheck`, `fatcontext`, `noctx`, `bodyclose`, `durationcheck`), and modernization (`perfsprint`, `usestdlibvars`, `intrange`, `copyloopvar`, …) linters; bumped `golangci-lint` to v2.13.0 and `goimports` to v0.49.0.
- CI and CodeQL workflows now build and test against Go 1.27.
- Tidied the `examples/` modules: `go mod tidy` refreshed their indirect dependencies (notably the AWS SDK in `examples/sns-provider`), and the `replace`-redirected `require` directives on the local modules were normalized to a single consistent version.

### Added

- Greptile code-review configuration under `.greptile/` (root `config.json`, `rules.md`, `files.json`) with a relaxed `examples/` override.

### Removed

- Go Report Card badge from the README — the service is no longer operational.

### Fixed

- `doc.go` package overview now lists the MSG91 provider alongside Twilio, SNS, and Vonage.

## [0.2.1] - 2026-06-07

### Changed

- SNS provider now wraps API errors from `SetSMSAttributes`, `GetSMSAttributes`, `CheckIfPhoneNumberIsOptedOut`, `ListPhoneNumbersOptedOut`, and `OptInPhoneNumber` with `gosms.ErrProviderError` for consistent error handling
- `MockProvider.Send` now reports a realistic segment count via `CalculateSegments` instead of a hardcoded `1`
- Re-enabled the `exported` revive linter and documented previously-undocumented exported symbols across the core and provider packages

### Fixed

- `MockProvider` ID generation now surfaces `crypto/rand` read failures instead of silently ignoring them
- Guarded `MultiProvider` strategy access with a mutex, preventing a data race when `WithStrategy` is called concurrently with `Send`
- `examples/twilio-provider` no longer reads the delivery status from a `nil` result when a scheduled send fails
- Example modules (`sns`, `twilio`, `vonage`) now require `gosms@v0.2.0` instead of the stale `v0.1.0`
- `make release-prep` now validates that `VERSION` starts with `v` and includes the `msg91` tag in its instructions

## [0.2.0] - 2026-04-21

### Added

- `gosms.WebhookParser` function type formalizing the `ParseWebhook` convention shared by all provider packages
- `gosms.OTPProvider` optional capability interface (`SendOTP`, `VerifyOTP`, `ResendOTP`) with shared `OTPRequest`, `OTPResult`, and `VerifyResult` types
  - MSG91 provider implements the interface, including a dedicated `/api/v5/otp` send path (server-side or client-supplied codes, template vars via `extra_param`)
- `msg91.Config.MaxRecipientsPerCall` caps how many recipients `SendBulk` packs into a single Flow API call (default 1000; negative disables chunking) so large batches are split automatically instead of exceeding MSG91's per-call limit
- MSG91 `ParseWebhook` falls back to MSG91's numeric `statusCode` when the textual `status` field is missing or non-standard
- `examples/webhooks` adds an MSG91 delivery-report handler alongside Twilio and Vonage
- `examples/msg91-provider` demonstrates `SendOTP`, `VerifyOTP`, and `ResendOTP`
- README documents MSG91 bulk chunking, `Body`→`body` template fallback, phone normalization, and per-message overrides

### Changed

- **BREAKING (msg91):** `Provider.RetryOTP` renamed to `Provider.ResendOTP` to match `gosms.OTPProvider`
- **BREAKING (msg91):** `Provider.VerifyOTP` now returns `*gosms.VerifyResult`

### Removed

- **BREAKING (msg91):** `msg91.VerifyResult` — use `gosms.VerifyResult`

### Fixed

- `make coverage` now includes the `msg91` module (was previously omitted from the merged profile)

## [0.1.1] - 2026-04-20

### Added

- MSG91 provider (`github.com/KARTIKrocks/gosms/msg91`) targeting the Flow API (v5) for DLT-compliant sending in India
  - Template-driven messaging via `SetVar` and `SetTemplateID` on `Message.Metadata`
  - Bulk send groups recipients by template + sender into a single API call
  - Provider-specific `VerifyOTP` / `RetryOTP` methods (not on the `Provider` interface)
  - `ParseWebhook` for MSG91 delivery report callbacks
  - Transactional and promotional route support
- `lint-fix` Makefile target running `golangci-lint run --fix` across all modules
- README section documenting the MSG91 provider

## [0.1.0] - 2026-03-17

### Added

- Core `Provider` interface with `Send`, `SendBulk`, `GetStatus`, and `Name` methods
- `Client` with default sender support and message validation
- `Message` builder with fluent API (`WithFrom`, `WithSchedule`, `WithMetadata`, etc.)
- `MultiProvider` with fallback and round-robin strategies
- `MockProvider` for unit testing with configurable behavior
- `SendEach` helper to deduplicate bulk-send logic across providers
- Phone number validation (`ValidateE164`) and normalization (`NormalizePhone`)
- SMS segment calculation with proper GSM 03.38 charset support
- Convenience helpers: `OTPMessage`, `AlertMessage`, `NotificationMessage`, `Batch`, `SendToMany`
- Twilio provider (`github.com/KARTIKrocks/gosms/twilio`) with webhook parsing
- AWS SNS provider (`github.com/KARTIKrocks/gosms/sns`) with opt-out management
- Vonage provider (`github.com/KARTIKrocks/gosms/vonage`) with delivery receipt parsing
- Multi-module architecture: providers are separate Go modules to avoid pulling unnecessary dependencies
- CI workflows (test, lint, coverage, CodeQL, benchmarks)
- Dependabot configuration for all modules
