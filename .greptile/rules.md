# Style and pattern rationale

Context for the scoped rules in `config.json`. This file is freeform prose
read alongside the diff; `config.json` is what actually gates comment scope
and severity.

## Provider parity, and why SNS was the outlier

CHANGELOG 0.2.1 describes fixing SNS so it wraps API errors from
`SetSMSAttributes`, `GetSMSAttributes`, `CheckIfPhoneNumberIsOptedOut`,
`ListPhoneNumbersOptedOut`, and `OptInPhoneNumber` in `gosms.ErrProviderError`
"for consistent error handling" — implying Twilio, Vonage, and MSG91 already
did this and SNS had quietly drifted. That's the shape of bug the
`provider-parity` and `sentinel-error-mapping` rules exist to catch: a new
SDK call added to one provider that returns its raw error type instead of
going through the same wrapping every sibling provider already uses. When
reviewing a provider file, check what the other three providers do for the
same kind of operation (send, bulk send, status lookup, webhook parse) before
accepting a provider-specific shortcut.

## Concurrent state needs symmetric locking, not just a setter that locks

`MultiProvider.WithStrategy` used to set `p.strategy` under a mutex while
`Send` read it directly, unguarded — a classic asymmetric-locking data race
that `go test -race` would only catch if the interleaving happened to occur
during a run. The fix guards both the write and the read with the same `mu`.
When reviewing any new mutable field on `Client`, `MultiProvider`, or a
provider struct that's touched from `Send`/`SendBulk`, verify every access —
not just the ones near a `Lock()` call you can already see — goes through the
same synchronization.

## crypto/rand errors are not optional to check

`MockProvider.generateID` originally called `rand.Read` and used the buffer
regardless of whether the read succeeded, so a failure (rare, but possible —
e.g. an exhausted entropy pool in a constrained container) produced IDs built
from a partially-zeroed or fully-zero buffer instead of surfacing an error.
The fix propagates the read error. Any new code generating IDs, tokens, or
nonces via `crypto/rand` should check the error the same way `generateID`
does now, not the way it used to.

## Root module dependency purity

AGENTS.md is explicit: the root module (`sms.go`, `helpers.go`, `mock.go`)
has **no third-party dependencies** — that's the entire reason providers live
in separate modules (`twilio/`, `sns/`, `vonage/`, `msg91/`), each pulling in
only its own SDK. A PR that adds an import outside the standard library to a
root-level `.go` file is very likely misplaced, not a design decision to wave
through.

## Multi-module boundary

Five Go modules (root + four providers) tied together at dev time by the
committed `go.work`. A change to `sms.go`'s `Provider`/`OTPProvider`
interfaces, `Message`/`Result`/`Status` shapes, or a sentinel error needs the
matching change in whichever provider modules implement the affected surface,
in the same PR — the root module's own tests won't catch a mismatch, since
each provider module only sees the root version pinned in its own `go.mod`
`require` (overridden locally by `go.work`, not by a `replace`; see AGENTS.md's
per-module tagging process for why the release order — root tag, then bump
each provider's require, then provider tags — matters).

## Shared helpers exist so providers don't quietly diverge

`ValidateE164`/`NormalizePhone` and `CalculateSegments`/`IsGSMEncoding`/
`GSMLen` live in root `helpers.go` precisely so every provider computes phone
validity and segment/cost estimates the same way. A provider that inlines its
own regex or septet-counting logic instead of calling these will drift the
moment either helper is tuned (e.g. a GSM charset fix), and the drift is easy
to miss because it only shows up as a slightly different cost or rejection
rate per provider, not a compile error or an obvious test failure.
