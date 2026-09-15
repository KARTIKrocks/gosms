---
id: helpers
title: Helpers
description: Phone validation and normalization, SMS segment counting, and message templates.
---

# Helpers

Provider-agnostic utilities in the core module. None of these touch the
network.

## Phone number utilities

```go
// Validate E.164 format.
if gosms.ValidateE164("+15551234567") {
    log.Println("Valid E.164 number")
}

// Normalize a phone number, prefixing a default country code if missing.
normalized := gosms.NormalizePhone("555-123-4567", "+1")
// Returns: +15551234567
```

`NormalizePhone` strips everything but digits (and a leading `+`), then
prefixes `defaultCountryCode` when the result doesn't already start with `+`.
It's a basic normalization and may not work for every country's numbering
plan — validate with `ValidateE164` afterward if correctness matters.

## SMS segment calculation

```go
// Check if a message fits GSM 7-bit encoding.
if gosms.IsGSMEncoding("Hello world") {
    log.Println("GSM encoding (160 char limit)")
}

// Calculate how many SMS segments a message will use.
segments := gosms.CalculateSegments("Hello, this is a test message!")
```

`CalculateSegments` implements the real GSM 03.38 / UCS-2 concatenation rules:

- **GSM 7-bit**: 160 septets single-segment, 153 septets per segment once
  concatenated. Extended characters (`{`, `}`, `[`, `]`, `~`, `\`, `^`, `|`,
  `€`) count as 2 septets each.
- **Unicode (UCS-2)**, used when the message contains any character outside
  the GSM alphabet: 70 UTF-16 code units single-segment, 67 per segment once
  concatenated. Characters outside the Basic Multilingual Plane (most emoji)
  count as 2 code units.

`GSMLen` and `IsGSMEncoding` are exported separately if you need the septet
count or the encoding check on their own.

## Pre-built message templates

```go
// OTP: "123456 is your MyApp verification code."
msg := gosms.OTPMessage("+15551234567", "123456", "MyApp")

// Alert: "[URGENT] Server is down!"
msg := gosms.AlertMessage("+15551234567", "URGENT", "Server is down!")

// Notification: "Order Update: Your order has shipped"
msg := gosms.NotificationMessage("+15551234567", "Order Update", "Your order has shipped")
```

Each helper also sets `Message.Metadata` (`type`, plus a template-specific
key) so consumers can filter by message kind downstream.

## `QuickSend`

A one-line convenience wrapping `NewClient(...).WithDefaultFrom(...).Send(...)`
for scripts and one-off sends:

```go
result, err := gosms.QuickSend(ctx, provider, "+15559876543", "+15551234567", "Hello!")
```
