---
id: sns
title: AWS SNS
description: Configuring the AWS SNS provider and its account-level extras.
---

# AWS SNS

```bash
go get github.com/KARTIKrocks/gosms/sns
```

## Configuration

Unlike the other providers, `sns.NewProvider` takes a `context.Context` —
the underlying AWS SDK client is constructed during setup:

```go
config := sns.DefaultConfig()
config.Region = "us-east-1"
config.AccessKeyID = "access_key"
config.SecretAccessKey = "secret_key"
config.SenderID = "MyApp"
config.SMSType = sns.SMSTransactional

provider, err := sns.NewProvider(ctx, config)
if err != nil {
    log.Fatal(err)
}

client := gosms.NewClient(provider)
result, err := client.Send(ctx, "+15559876543", "Your code is 123456")
```

| Field | Purpose |
| --- | --- |
| `Region` | AWS region |
| `AccessKeyID`, `SecretAccessKey` | AWS credentials (omit to use the default credential chain) |
| `SenderID` | Alphanumeric sender ID, where supported by the destination country |
| `SMSType` | `sns.SMSTransactional` or `sns.SMSPromotional` |
| `MaxPrice` | Optional per-message spend cap |
| `Client` | Optional — supply a pre-configured `*sns.Client` instead of building one from the fields above |

## Account-level extras

SNS exposes account-level operations beyond the `Provider` interface —
call these on the concrete `*sns.Provider`, not through `gosms.Client`:

```go
provider, _ := sns.NewProvider(ctx, config)

// Set account-level SMS attributes.
err := provider.SetSMSAttributes(ctx,
    "100.00",                         // Monthly spend limit
    "arn:aws:iam::123:role/SNSRole",  // IAM role for delivery logs
    "100",                            // Success sampling rate %
)

// Opt-out management.
optedOut, err := provider.CheckIfPhoneNumberIsOptedOut(ctx, "+15551234567")
numbers, err := provider.ListPhoneNumbersOptedOut(ctx)
err = provider.OptInPhoneNumber(ctx, "+15551234567")
```

## Delivery status

SNS's `GetStatus` returns `gosms.ErrUnsupported` — SNS does not offer a
message-status lookup API. Use `SetSMSAttributes` to configure delivery
status logging to CloudWatch instead, or configure a webhook where the
destination platform supports one.
