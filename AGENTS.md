# Repository Rules

## Documentation Requirements

### `decision.md` Is Mandatory

- Core decisions must be recorded in [decision.md](/Users/erpang/GitHubProjects/post-sync/decision.md).
- Any work that changes architecture, dependency shape, data flow, rendering strategy, deployment, channel semantics, or operational assumptions must update `decision.md` in the same round.

### `design.md` Must Be Concrete

`design.md` is not allowed to stay at the level of generic discussion. It must contain at least:

- Clear module boundaries
- Clear data models
- Clear API design
- Clear state design
- Clear dedup design
- Clear template design
- Clear deployment design

If a change makes any of those sections stale, update [design.md](/Users/erpang/GitHubProjects/post-sync/design.md).

### `decision.md` Must Explicitly Record

- All newly added dependencies
- All key architectural decisions
- All key assumptions

## Commit Rules

### General Principle

All changes must follow the principle of small modules, low coupling, and easy review.

### Commit Ordering

- Do not combine changes that can be committed independently.
- Prefer committing non-behavioral groundwork first, such as:
  - directory structure
  - type definitions
  - interface definitions
  - documentation
- Then commit behavioral changes, such as:
  - repository
  - service
  - handler
  - adapter
  - page
- If documentation updates can stand alone, commit them separately.

### Commit Messages

Each commit message must use an imperative form or a clear action, for example:

- `add publish job model`
- `add telegram channel adapter interface`
- `add markdown dedup normalization`
- `document deployment flow`

Avoid vague summaries that hide the main action.

## Working Expectations

- When in doubt, document the decision.
- When in doubt, split the commit.
- If a change cannot be reviewed independently, it is probably too large for one commit.
