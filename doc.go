// Package gosms provides a unified, provider-agnostic interface for sending
// SMS messages in Go.
//
// # Architecture
//
// The library is organized as a multi-module repository to keep dependencies
// minimal. The core module defines the shared types and interfaces with zero
// external dependencies. Each provider lives in its own sub-module so users
// only download the dependencies they actually need:
//
//   - github.com/KARTIKrocks/gosms          — core types, client, helpers, mock
//   - github.com/KARTIKrocks/gosms/twilio   — Twilio provider
//   - github.com/KARTIKrocks/gosms/sns      — AWS SNS provider
//   - github.com/KARTIKrocks/gosms/vonage   — Vonage (Nexmo) provider
//
// # Quick Start
//
// Send an SMS via Twilio:
//
//	import (
//	    "github.com/KARTIKrocks/gosms"
//	    "github.com/KARTIKrocks/gosms/twilio"
//	)
//
//	provider, err := twilio.NewProvider(twilio.Config{
//	    AccountSID: "AC...",
//	    AuthToken:  "...",
//	    From:       "+15551234567",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	client := gosms.NewClient(provider)
//	result, err := client.Send(ctx, "+15559876543", "Hello from gosms!")
//
// # Multi-Provider Fallback
//
// Use MultiProvider to try multiple providers in order:
//
//	multi := gosms.NewMultiProvider(twilioProvider, snsProvider)
//	client := gosms.NewClient(multi)
//
// Or rotate across providers with round-robin:
//
//	multi := gosms.NewMultiProvider(twilioProvider, snsProvider).
//	    WithStrategy(gosms.StrategyRoundRobin)
//
// # Testing
//
// Use the built-in MockProvider for unit tests:
//
//	mock := gosms.NewMockProvider()
//	client := gosms.NewClient(mock)
//
//	result, err := client.Send(ctx, "+15551234567", "test")
//	// assert result, check mock.Messages(), etc.
package gosms
