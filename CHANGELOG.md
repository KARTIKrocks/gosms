# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
