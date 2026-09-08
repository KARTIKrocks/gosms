# Review guidelines for gosms

`gosms` is a unified, dependency-light SMS library for Go. Reviews should hold it
to the standard of a public library that other projects depend on.

## What matters most

- **Backward compatibility.** The exported API is stable. Renaming or removing an
  exported identifier, changing a signature, adding a method to an exported
  interface, or altering a sentinel error is a breaking change and must be called
  out — even when the PR title does not say "breaking".
- **Error handling.** Errors are wrapped with `%w` and carry context. Provider
  packages translate provider-specific failures into the core sentinel errors
  (`ErrInvalidPhone`, `ErrRateLimited`, `ErrProviderError`, `ErrInvalidConfig`,
  …) so callers can use `errors.Is`. Returned errors are never dropped.
- **Context and resources.** Every network call takes a `context.Context` and
  honours cancellation and deadlines. HTTP response bodies are always closed.
- **Concurrency.** `Client`, `MockProvider`, `MultiProvider`, and all providers
  are documented safe for concurrent use. Shared state is guarded; new
  concurrent code comes with a `-race` test.
- **Secrets.** Credentials, tokens, and full phone numbers never reach logs,
  panics, or error strings.
- **Docs.** Every exported symbol has a doc comment starting with its name.

## The module layout

The repo is a Go workspace of five published modules — the root (core, **no
third-party deps**) plus one module per provider (`twilio`, `sns`, `vonage`,
`msg91`). A change to a core type usually needs matching changes in every
provider module. `go.work` handles local resolution; no published `go.mod`
carries a `replace` directive. The `examples/` modules are never published and
are allowed to use `replace`.

## Tone

Be concise and technical. Prioritise correctness, compatibility, and Go idioms
over stylistic nits. CI (`golangci-lint` v2, `go test -race`, CodeQL) is the
merge gate; Greptile findings are advisory.

## Go version

The project targets **Go 1.27** across every module. New code should use current
standard-library facilities where they fit — for example `testing/synctest` for
time-dependent tests, `min`/`max` and integer `range`, and `slices`/`maps`
helpers instead of hand-rolled loops.
