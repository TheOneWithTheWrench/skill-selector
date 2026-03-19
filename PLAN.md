# skill-switcher-v2 plan

## Purpose
- Rebuild `skill-switcher` from scratch with clearer boundaries, better tests, and code we both understand.
- Keep the product direction from v1, but treat v1 as a behavior reference rather than something to port file-by-file.
- Make the codebase good enough to open source.

## Agreed
- v2 is a rebuild/refactor of v1, not a new product.
- There must be one UI-independent core for source management, cataloging, profiles, and sync.
- CLI and TUI are separate entries on top of that core.
- Package boundaries should be deliberate; we only split packages around real concepts.
- The code style should take inspiration from `go-fly`.
- The test style should take inspiration from `go-liga`.
- `PLAN.md` is the living place for decisions, learnings, and TODOs.

## Collaboration agreement
- Build v2 in small, understandable slices.
- Prefer code we can explain in one pass over clever abstractions.
- When we introduce a new package or abstraction, capture why it exists here.
- Keep updating this file as we learn more from v1 and from the rewrite.
- Treat domain language as part of the work, not cleanup for later.

## Working principles
- Prefer small, cohesive packages over one large `internal` package.
- Do not let core packages depend on Bubble Tea, CLI formatting, or view models.
- Keep filesystem, `gh`, `git`, and symlink side effects behind narrow interfaces.
- Prefer explicit data types and functions over generic plumbing.
- Use Cobra for the CLI now that the command surface justifies it, keep commands as thin `RunE` wrappers around the shared app layer, and use Fang as a presentation layer for help and errors rather than pushing styling into the core.
- Prefer domain names over vague infrastructure names like `state` and `store` when a sharper name exists.
- Keep file versions and JSON schemas in repository code, not in the core entities.
- Let entities own validation and pure derivations; let services and repositories own side effects.
- Prefer clear naming over comments.
- Keep exported APIs small and documented.

## What v1 got right
- GitHub tree URL sources are a good fit for the problem.
- Source normalization and explicit file versioning are worth keeping.
- Path-based skill identity is simple and good enough for now.
- Symlink sync with owned manifests is the right foundation.
- Deduplicating adapters that share a root path is a good idea.
- The default roots for Claude, Opencode, Ampcode, Codex, and Cursor already match real usage and are worth carrying forward.

## Learnings from v1
- The current `internal` package does too much orchestration and owns too many concepts.
- App logic leaks TUI types, which couples the core to one interface.
- Domain types are duplicated across packages, which creates avoidable mapping code.
- Names like `state` and `store` are too generic and hide the actual domain boundaries.
- The TUI model is too large; it should be split into smaller focused pieces once the core API is stable.
- Pure logic and side effects are mixed too often; v2 should separate planning from mutation where possible.
- We should optimize for readable code and tests, not fastest possible shipping.
- TUI dirty state is interface-local state, not domain state. The core should accept a desired selection; the TUI should own the draft that differs from what is currently active.

## Domain language
- `Source` means a configured upstream skill source with a stable ID, a user-facing locator, a fetch URL, a ref, and a subtree.
- `Sources` means the normalized collection of configured sources and owns rules like add, remove, deduplicate, and stable ordering.
- `Sources` should be a named slice type (`type Sources []Source`) so empty collections stay idiomatic and error returns can use `nil`.
- `Repository` means a persistence boundary. It is infrastructure, not a domain entity.
- `Mirror` means the managed local clone of a `Source`. It is not the source itself.
- Clone and pull behavior should live in a source refresh service, not on `Source`.
- `Refresher` means the service that materializes or updates `Mirror` instances from their upstream `Source` definitions.
- Provider-specific parsing should live behind source package helpers. We do not need parser interfaces until we have more than one real provider.
- `Skill` means one discovered skill directory under a mirrored source subtree.
- `Skill` should own a `SkillIdentity` plus catalog metadata, rather than duplicating `SourceID` and relative path fields.
- `Skills` means the normalized collection of discovered skills.
- `Catalog` means the indexed inventory of discovered skills plus when that inventory was generated.
- Catalog scanning should live in an explicit scanner service or function, not on `Mirror` or `Source`.
- `SkillIdentity` means a lightweight identity for a skill by `SourceID` and relative path only; it must not carry remote, catalog, or presentation metadata.
- `Target` means one sync destination with a root path and link mapping strategy.
- `Manifest` means the persisted set of `SkillIdentity` values currently owned by one sync target.
- Sync should operate on `SkillIdentity`, not `catalog.Skill`, so profiles can later reuse the same selection model without depending on catalog metadata.
- The same rule should apply in other slices: keep entities pure, keep side effects in explicit services.

## v1 MVP behavior to preserve
- Manage sources as GitHub tree URLs stored in a versioned JSON file.
- Generate a stable source ID from the repo, ref, and subtree.
- Clone or pull each source into a managed local sources directory.
- Scan the chosen subtree for skill directories that contain `SKILL.md`.
- Persist a catalog snapshot built from discovered skills.
- Support multiple profiles with one active profile and persisted selected skills.
- Sync the active selected skills into supported agent folders with symlinks.
- Persist sync manifests so later pulls and switches can reconcile existing links.
- Support both CLI and TUI entrypoints on top of the same core behavior.

## Architecture direction
- Treat v1 as a behavior reference, not a structural reference.
- Build a core application layer that exposes use cases such as:
  - add, remove, and list sources
  - refresh sources
  - scan and build the catalog
  - list and select skills for the active profile
  - switch profiles
  - plan and apply sync
  - inspect current sync state
- Keep CLI and TUI thin:
  - CLI parses args, calls the core, and prints text
  - TUI manages state/rendering, calls the core, and renders view state
- Return core/domain results from the application layer. CLI and TUI should map those results into presentation-specific models locally.
- Rebuild the TUI after the core and CLI have made the boundaries real.
- The TUI should treat sync manifests as the current active selection and hold a separate session-local desired selection. Sync should stay explicit, and quitting the TUI should drop unsynced draft changes.
- As we rebuild each slice, we should stop and name the entities before copying behavior from v1.

## First pass package boundaries
- `cmd/skill-switcher/` - process entrypoint
- `internal/app/` - orchestration and shared use cases
- `internal/cli/` - command parsing and text output
- `internal/tui/` - Bubble Tea program and view models
- `internal/source/` - `Source`, `Sources`, source repository, local mirrors, and later refresh/update services
- `internal/catalog/` - `Skill`, `Skills`, `Catalog`, scanning, and catalog repository
- `internal/skillidentity/` - lightweight skill identities shared by sync and later profiles
- `internal/sync/` - targets, manifests, reconciliation, and sync persistence
- `internal/profile/` - profiles, selection state, persistence
- `internal/agent/` - supported agent adapters and detection
- `internal/paths/` - runtime paths and XDG locations

Package rule:
- create a package when the concept has its own model, behavior, and tests
- otherwise keep it local instead of abstracting early

## Testing style
- Use `go-fly` as the bar for package shape and focused tests.
- Use `go-liga` as the bar for helper patterns and assertion quality.
- Prefer `moq` with a package-local `gen.go` and generated `mocks_test.go` over handwritten mock structs when the dependency is an interface.
- Prefer local `dependencies` structs plus helpers like `newSut`, `newCtx`, and `newDefaultDependencies`.
- Give `newDefaultDependencies` sensible happy-path behavior, then override only the dependency behavior relevant to each test case.
- Assert on expected `moq` calls explicitly, and assert zero calls for dependencies that should not be touched.
- Use `require.Len(...Calls(), 1)` before indexing into generated `moq` call records.
- Keep failure cases before happy path.
- Prefer blackbox package tests where practical.
- Use temp dirs for filesystem tests and test doubles for `gh`, `git`, and OS interactions.
- Keep the core heavily unit-tested before spending much time on TUI behavior tests.
- Done means `go test ./... -race`, `go vet ./...`, `go fmt`, and `goimports` pass.

## Recommended build order
1. Lock down the v1 behavior we want to keep.
2. Scaffold the v2 module, paths, and shared test helpers.
3. Implement source parsing, persistence, and fetch flow.
4. Implement catalog scanning.
5. Implement sync around lightweight skill identities and manifests.
6. Implement application use cases on top.
7. Implement CLI on top of the core.
8. Implement profile and selection state once the sync and selection model is stable.
9. Implement TUI on top of the same core.
10. Polish docs, examples, and OSS readiness.

## TODO
- [x] Create this living plan document.
- [x] Review v1 and write down the exact MVP behavior we want to preserve in v2.
- [x] Decide first-pass v2 package names for the initial slice.
- [x] Clarify the source slice domain language and entity responsibilities.
- [x] Clarify the catalog slice domain language and entity responsibilities.
- [x] Initialize the Go module in `skill-switcher-v2/`.
- [x] Add a small `cmd/skill-switcher` entrypoint that only wires dependencies.
- [x] Build the shared application layer with no CLI/TUI imports.
- [x] Move source parsing and source persistence into their own package.
- [x] Move catalog discovery into its own package.
- [x] Move agent adapter and sync logic into focused packages.
- [x] Implement a first end-to-end CLI flow against the shared core.
- [x] Add a partial TUI for sources, catalog browsing, draft selection, and sync without pulling profile logic back into the core.
- [ ] Move profile logic into its own package after the sync and selection model settles.
- [ ] Extend the TUI to full v1 parity once the profile slice exists.
- [ ] Add high-quality package tests across core concepts.
- [ ] Write README and OSS-facing docs once the structure settles.

## Open questions
- What should the public repo/module name be when we open source it?
- Do we keep the full current adapter set from day one, or start with a smaller set and add the rest back?
- Do we keep GitHub tree URLs as the only source type in the v2 MVP?
- Should sync continue to apply immediately on profile switch, or should that become an explicit action everywhere?
- How much of the current TUI UX do we want to preserve versus simplify before rebuilding?
