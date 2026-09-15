---
id: otp
title: OTP
description: The optional OTPProvider capability for send / verify / resend flows.
---

# OTP

`OTPProvider` is an optional capability for providers that expose a dedicated
OTP flow, rather than requiring callers to assemble OTP messages through the
generic SMS path. Providers without native OTP support simply don't implement
it — callers detect the capability with a type assertion. Currently only
[MSG91](msg91) implements it.

```go
type OTPProvider interface {
    SendOTP(ctx context.Context, req *OTPRequest) (*OTPResult, error)
    VerifyOTP(ctx context.Context, phone, otp string) (*VerifyResult, error)
    ResendOTP(ctx context.Context, phone, channel string) error
}
```

## Usage

```go
if otp, ok := provider.(gosms.OTPProvider); ok {
    // MSG91 generates the code server-side when OTPRequest.OTP is empty.
    _, err := otp.SendOTP(ctx, &gosms.OTPRequest{
        Phone:  "+919876543210",
        Length: 6,
        Expiry: 5 * time.Minute,
    })

    vr, err := otp.VerifyOTP(ctx, "+919876543210", "123456")
    if err == nil && vr.Verified {
        // OTP matched
    }

    // Resend via "text" or "voice".
    _ = otp.ResendOTP(ctx, "+919876543210", "voice")
}
```

## Types

`OTPRequest` (passed to `SendOTP`):

| Field | Purpose |
| --- | --- |
| `Phone` | Recipient phone number (E.164 recommended) |
| `OTP` | Code to send; when empty, providers that support server-side generation produce one |
| `Length` | Desired OTP length when the provider generates the code (ignored when `OTP` is set); typical range 4-9 |
| `Expiry` | How long the OTP remains valid |
| `TemplateID` | Provider template identifier (e.g. a DLT template ID for MSG91); overrides any default configured on the provider |
| `Vars` | Template variables beyond the OTP itself |

`OTPResult` (returned by `SendOTP`): `MessageID`, `Phone`, `SentAt`, `Raw`.

`VerifyResult` (returned by `VerifyOTP`): `Verified bool`, `Message string`
(raw provider message, useful for failure reasons), `Raw map[string]any`.
