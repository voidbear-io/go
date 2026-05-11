# Agent notes

## Commit strategy

Use **scoped prefixes** so history stays readable across this repo (multiple packages may live here).

### Work in progress

While iterating locally or before a change is ready for review, prefix commits with **`wip`** and the **scope** (area or module):

```text
wip(cmdargs): short imperative description
```

Examples:

- `wip(cmdargs): add GetMap tests`
- `wip(cmdargs): fix Split edge case`

Rules:

- **Scope** — Use `cmdargs` for the `cmdargs` module; pick another scope (e.g. `ci`, `repo`) when the change is not under `cmdargs/`.
- **Subject** — Imperative mood, lowercase after the colon, no trailing period, ~72 characters or less when possible.
- **Push** — WIP commits may be pushed to a personal branch; squash or reword before opening a PR if history should stay clean.

### Ready to merge

For commits intended to land on the main branch as-is, drop the `wip` prefix and keep the same scope style when it helps:

```text
cmdargs: describe the finalized change
```

or follow your team’s preferred convention (e.g. Conventional Commits: `feat(cmdargs): …`, `fix(cmdargs): …`).
