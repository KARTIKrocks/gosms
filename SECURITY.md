# Security Policy

## Supported versions

gosms is a multi-module repository — the root module plus one module per
provider (`twilio`, `sns`, `vonage`, `msg91`), each tagged independently (see
[AGENTS.md](AGENTS.md#releasing-per-module-tags)). Only the **latest tagged
version of each module** is supported with security fixes. There is no
long-term-support branch; upgrading to the latest tag is the fix.

## Reporting a vulnerability

**Do not open a public issue for a suspected vulnerability.** Report it
privately using [GitHub's private vulnerability reporting](https://github.com/KARTIKrocks/gosms/security/advisories/new)
for this repository.

Include, where possible:

- The affected module(s) and version(s).
- A minimal reproduction or proof of concept.
- The impact you believe the issue has (e.g. credential exposure, request
  forgery, denial of service).

You should expect an initial response within a few days. Once a fix is
available, it will be released and credited in the [changelog](CHANGELOG.md)
and a GitHub Security Advisory, unless you request otherwise.

## What's in scope

- The core module and all four provider modules.
- Handling of credentials and secrets passed into provider constructors
  (these must never be logged — see [AGENTS.md](AGENTS.md#conventions)).
- Webhook parsing (`ParseWebhook` in the `twilio`, `vonage`, and `msg91`
  packages) — a malformed or malicious request body should error, never panic
  or execute untrusted content.

Vulnerabilities in a third-party SMS provider's own API or infrastructure are
out of scope here — report those to the provider directly.

## What gosms's CI already checks

Every push and pull request to `main` is scanned by
[CodeQL](https://github.com/KARTIKrocks/gosms/actions/workflows/codeql.yml),
with a full re-scan weekly to catch newly published query patterns against
unchanged code. [Dependabot](.github/dependabot.yml) tracks updates across
every Go module, the GitHub Actions themselves, and the documentation site's
npm dependencies.
