---
id: providers
title: Providers Overview
description: The four supported SMS backends and how to choose between them.
---

# Providers Overview

Each provider is its own Go module implementing the core [`Provider`
interface](core#provider-interface). Install only the ones you use.

| Provider | Module | Best for |
| --- | --- | --- |
| [Twilio](twilio) | `github.com/KARTIKrocks/gosms/twilio` | Global SMS, scheduled sends, webhook delivery receipts |
| [AWS SNS](sns) | `github.com/KARTIKrocks/gosms/sns` | Teams already on AWS; transactional/promotional SMS type, opt-out management |
| [Vonage](vonage) | `github.com/KARTIKrocks/gosms/vonage` | Global SMS with webhook delivery receipts |
| [MSG91](msg91) | `github.com/KARTIKrocks/gosms/msg91` | India — DLT-approved Flow templates, server-side OTP generation |

All four satisfy the same `Provider` interface, so switching providers (or
running several behind a [`MultiProvider`](multi-provider)) never changes
call-site code beyond construction:

```go
provider, err := twilio.NewProvider(twilio.Config{ /* ... */ })
// or: sns.NewProvider(ctx, sns.DefaultConfig())
// or: vonage.NewProvider(vonage.Config{ /* ... */ })
// or: msg91.NewProvider(msg91.Config{ /* ... */ })

client := gosms.NewClient(provider)
result, err := client.Send(ctx, "+15559876543", "Hello from gosms!")
```

## Capability matrix

| Capability | Twilio | SNS | Vonage | MSG91 |
| --- | --- | --- | --- | --- |
| Free-text `Body` | ✓ | ✓ | ✓ | Fallback only — [templated](msg91#template-variables-and-body-fallback) |
| Scheduled sends (`WithSchedule`) | ✓ | – | – | – |
| Webhook `ParseWebhook` | ✓ | – | ✓ | ✓ |
| `OTPProvider` | – | – | – | ✓ |
| Account-level extras | – | Opt-out mgmt, SMS attributes | – | – |

See each provider's page for configuration fields and provider-specific
notes.
