# Skill Selector

[![CI](https://github.com/TheOneWithTheWrench/skill-selector/actions/workflows/ci.yml/badge.svg)](https://github.com/TheOneWithTheWrench/skill-selector/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/actions/workflow/status/TheOneWithTheWrench/skill-selector/release.yml?label=release)](https://github.com/TheOneWithTheWrench/skill-selector/actions/workflows/release.yml)
[![GitHub release](https://img.shields.io/github/v/release/TheOneWithTheWrench/skill-selector)](https://github.com/TheOneWithTheWrench/skill-selector/releases)
[![Go version](https://img.shields.io/github/go-mod/go-version/TheOneWithTheWrench/skill-selector)](https://go.dev/doc/devel/release)

🧰 Skill Selector manages shared skills across multiple coding agents from one place.

It lets you:
- add upstream skill sources from GitHub tree URLs
- build a local catalog of discovered skills
- save named profiles of selected skills
- sync those profiles into supported agents with symlinks
- work from either a TUI or a CLI on top of the same core

## ✨ Supported Agents (But adding more is easy)

- Ampcode
- Claude
- Codex
- Cursor
- Opencode

## 🧠 Mental Model

- `Source` - an upstream skill source, currently a GitHub tree URL
- `Catalog` - the local inventory of discovered skills from refreshed sources
- `Profile` - a named saved selection of skills
- `Sync` - the real installed symlinks in agent skill directories

High-level flow:
1. add one or more sources
2. refresh to build the catalog
3. select skills into a profile
4. sync that profile into your agents

## 🚀 Install

With Homebrew:

```bash
brew install TheOneWithTheWrench/tap/skill-selector
```

From this repo:

```bash
go install ./cmd/skill-selector
```

Or run it directly without installing:

```bash
go run ./cmd/skill-selector
```

If your Go bin directory is on `PATH`, the installed command is:

```bash
skill-selector
```

## 🎛️ TUI

The TUI is the default interface:

```bash
skill-selector
```

You can also open it explicitly:

```bash
skill-selector tui
```

The main sections are:
- `Sources` - manage upstream sources and drill into one source's skills
- `Catalog` - toggle skills for the active profile
- `Profiles` - create, rename, remove, and activate profiles
- `Status` - inspect runtime paths and current sync state

## 💻 CLI

Use the CLI when you want direct commands, scripting, or quick inspection.

### Sources

Add a source:

```bash
skill-selector source add https://github.com/anthropics/skills/tree/main/skills
```

List configured sources:

```bash
skill-selector source list
```

Remove a source by locator or source ID:

```bash
skill-selector source remove https://github.com/anthropics/skills/tree/main/skills
skill-selector source remove anthropics-skills-skills-75224e3c
```

### Refresh

Refresh all source mirrors and rebuild the local catalog:

```bash
skill-selector refresh
```

Alias:

```bash
skill-selector pull
```

### Catalog

List discovered skills:

```bash
skill-selector catalog list
```

### Profiles

List profiles:

```bash
skill-selector profile list
```

Create a profile:

```bash
skill-selector profile create reviewer
```

Rename a profile:

```bash
skill-selector profile rename reviewer backend-reviewer
```

Remove an inactive profile:

```bash
skill-selector profile remove backend-reviewer
```

Activate a profile and sync it immediately:

```bash
skill-selector profile switch reviewer
```

### Sync

Sync one or more explicit skill identities:

```bash
skill-selector sync source-id:reviewer
skill-selector sync source-id:path/to/skill
```

Sync the full catalog:

```bash
skill-selector sync --all
```

Inspect persisted sync manifests:

```bash
skill-selector sync status
```

Clear all managed synced skills:

```bash
skill-selector sync clear
```

## 🔄 Typical Workflow

Example setup:

```bash
skill-selector source add https://github.com/anthropics/skills/tree/main/skills
skill-selector refresh
skill-selector
```

Then in the TUI:
- choose or create a profile
- toggle the skills you want
- sync them into your agents

## 📁 Runtime Data

By default Skill Selector stores data in XDG-style paths:

- data: `~/.local/share/skill-selector`
- cache: `~/.cache/skill-selector`

That includes sources, profiles, sync manifests, and the local catalog cache.

## 🤝 Notes

- `skill-selector --help` shows the full command tree
- the CLI and TUI share the same core behavior
- the project intentionally keeps the core independent from Cobra and Bubble Tea
