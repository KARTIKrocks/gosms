# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
