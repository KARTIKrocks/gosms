---
id: multi-provider
title: Multi-Provider
description: Composing multiple providers behind one Provider with fallback or round-robin.
---

# Multi-Provider

`MultiProvider` is itself a `Provider` that wraps several providers behind
one interface, selecting between them with `StrategyFallback` or
`StrategyRoundRobin`. Compose one and pass it straight to `gosms.NewClient` —
call sites don't change.

## Fallback

Try providers in order until one succeeds. This is the default strategy:

```go
multi := gosms.NewMultiProvider(twilioProvider, vonageProvider)

client := gosms.NewClient(multi)
result, err := client.Send(ctx, to, body)
```

## Round-robin

Rotate across providers using an atomic counter:

```go
multi := gosms.NewMultiProvider(twilioProvider, vonageProvider).
    WithStrategy(gosms.StrategyRoundRobin)

client := gosms.NewClient(multi)
```

## Behavior notes

- `Name()` returns `"multi"`.
- `SendBulk` sends each message individually through `Send` (so the selected
  strategy applies per message), collecting results with `SendEach` — a
  provider error for one message never aborts the batch.
- `GetStatus` queries providers in order and returns the first successful
  lookup; if none succeed, it returns `gosms.ErrUnsupported`.
- `MultiProvider` is safe for concurrent use — the round-robin counter is a
  `sync/atomic.Uint64`, and the strategy itself is guarded by a `RWMutex`.
