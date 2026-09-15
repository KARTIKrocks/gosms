# Documentation versioning

gosms doesn't have a version snapshot yet — `versions.json` is `[]` and
everything lives in `docs/`, served at `/docs/`. The infrastructure below
exists so that the *first* release which changes documented behaviour has
somewhere to go without inventing the process under time pressure. Until
then, treat this file as the runbook you'll follow, not a description of
something already in place.

## The rules

### 1. Snapshot when a release changes documented behaviour — not on every release

A snapshot exists to answer one question: *what was true before this release
broke it?* If nothing broke, the snapshot is a byte-identical copy of every doc
page that has to be maintained forever.

**The test:** does this release make an existing page **wrong for someone on
the previous version**? A changed default, changed semantics, a rename, a
removal, a deprecation — those get a snapshot. Purely additive releases do
not (see rule 2).

Go makes this rule stronger than it would be elsewhere. Semver plus the Go
compatibility promise means additive minors leave the documented contract
intact, and a `v2` is a different import path
(`github.com/KARTIKrocks/gosms/v2`) with its own pkg.go.dev — a genuinely
different package. **Always snapshot at a major.**

### 2. Additive changes get a version marker, not a snapshot

The one real problem a reader on an older version has is the inverse of a
breaking change: they read about something that does not exist in their
version yet, and it does not compile. Snapshots are an expensive fix for
that. A marker is a cheap one, and it is a better answer — a snapshot tells
you what existed, a marker tells you what upgrading buys you.

The convention is plain Markdown, so it needs no components and survives
being copied into a snapshot:

| Situation | Write |
| --- | --- |
| New row in an API table | append `_0.2+_` to the description cell |
| New option or behaviour in prose | open the paragraph with `_Added in 0.2._` |
| New member inside a code block | trailing `// 0.2+` comment |
| Behaviour that changed | `_Changed in 0.3._` plus one line on what it was before |

Markers use `MAJOR.MINOR` — matching snapshot names — so they stay greppable.
Drop a marker once it names a version older than the oldest live snapshot; by
then everyone reading has it.

### 3. Snapshots are `MAJOR.MINOR`, never patch

Versions would be `1.0`, `1.1`, `2.0` — never `1.0.0`, never `v1.0`.

A patch release that changes documented behaviour is **edited into the
existing `versioned_docs/version-<minor>/` in place**. It does not get its own
snapshot. A patch by definition does not change the contract; if it did, the
docs were already wrong, and the fix belongs in the snapshot that is wrong.

`npm run cut-version` enforces the format; it rejects anything that isn't
`MAJOR.MINOR`, already exists, or is older than the current newest.

### 4. Only the newest 4 versions are built

`MAX_LIVE_VERSIONS` in `versions.config.json` caps how many snapshots get
built and indexed. Older ones stay in git — readable at their tag, restorable
by bumping the constant — but they don't cost build time or search index
size. This keeps build time flat as releases accumulate instead of growing
linearly. Raise it only if the reason is real support obligations, not
release count.

### 5. `docs/` is the future, not the present

| Directory | Serves | URL |
| --- | --- | --- |
| `docs/` | **Next** — unreleased, tracks `main` | `/docs/next/` (once a snapshot exists) |
| `versioned_docs/version-<minor>/` | the current release | `/docs/` |

Until the first snapshot is cut, `docs/` is served directly at `/docs/` — the
`current`/`next` split in `docusaurus.config.ts` only takes effect once
`versions.json` is non-empty. A PR that changes documented behaviour always
edits `docs/`, not a snapshot — the snapshot (once one exists) is frozen
history.

## Release runbook

Every release starts the same way, and then rule 1 decides whether it ends
there.

### Every release

Make sure `docs/` describes the release accurately — everything merged into
`main` since the last release should already be reflected there — and that
new APIs carry their version markers once a first snapshot exists (before
that, there's nothing to mark against).

### If the release only adds

Nothing else to do. `docs/` becomes the new truth on the next deploy.

### If the release changes documented behaviour

```bash
cd website
npm run cut-version -- 0.2
npm run check
```

`versions.json`, `versioned_docs/version-0.2/`, and
`versioned_sidebars/version-0.2-sidebars.json` are created for you, `/docs/`
starts serving 0.2, and `docs/` becomes `/docs/next/`.

### Patch releases

Never a snapshot. Edit `versioned_docs/version-<minor>/` directly (once one
exists), and mirror the change into `docs/` if it still applies to `main`.

## Day-to-day

```bash
npm start          # dev server
npm run build      # production build
npm run check      # lint + typecheck + build
```

## Adding a page

1. Create `docs/<id>.md` with `id`, `title`, and `description` frontmatter.
2. Add the id to `sidebars.ts` under the right category.

The sidebar is explicit rather than autogenerated so ordering is a deliberate
choice and a new file can't silently rearrange the nav.

## What does not belong here

Type signatures, method sets, and struct fields. Those are generated from
source on [pkg.go.dev](https://pkg.go.dev/github.com/KARTIKrocks/gosms) and
are always correct; hand-copying them here creates a second source of truth
that drifts.

These guides cover concepts, grouped overviews, configuration, and worked
examples — and link out for the exact signatures.
