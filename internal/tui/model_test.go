package tui

import (
	"context"
	"fmt"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-selector/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type workflowStub struct {
	addSourceFunc      func(context.Context, string) (SourceActionResult, error)
	removeSourceFunc   func(context.Context, string) (SourceActionResult, error)
	refreshFunc        func(context.Context) (RefreshActionResult, error)
	createProfileFunc  func(context.Context, string) (ProfilesActionResult, error)
	renameProfileFunc  func(context.Context, string, string) (ProfilesActionResult, error)
	removeProfileFunc  func(context.Context, string) (ProfilesActionResult, error)
	switchProfileFunc  func(context.Context, string) (ProfilesActionResult, error)
	syncFunc           func(context.Context, skill_identity.Identities) (SyncActionResult, error)
	addSourceCalls     []string
	removeCalls        []string
	refreshCalls       int
	createProfileCalls []string
	renameProfileCalls [][2]string
	removeProfileCalls []string
	switchProfileCalls []string
	syncCalls          []skill_identity.Identities
}

func (w *workflowStub) AddSource(ctx context.Context, locator string) (SourceActionResult, error) {
	w.addSourceCalls = append(w.addSourceCalls, locator)
	if w.addSourceFunc == nil {
		return SourceActionResult{}, nil
	}

	return w.addSourceFunc(ctx, locator)
}

func (w *workflowStub) RemoveSource(ctx context.Context, identifier string) (SourceActionResult, error) {
	w.removeCalls = append(w.removeCalls, identifier)
	if w.removeSourceFunc == nil {
		return SourceActionResult{}, nil
	}

	return w.removeSourceFunc(ctx, identifier)
}

func (w *workflowStub) Refresh(ctx context.Context) (RefreshActionResult, error) {
	w.refreshCalls++
	if w.refreshFunc == nil {
		return RefreshActionResult{}, nil
	}

	return w.refreshFunc(ctx)
}

func (w *workflowStub) CreateProfile(ctx context.Context, name string) (ProfilesActionResult, error) {
	w.createProfileCalls = append(w.createProfileCalls, name)
	if w.createProfileFunc == nil {
		return ProfilesActionResult{}, nil
	}

	return w.createProfileFunc(ctx, name)
}

func (w *workflowStub) RenameProfile(ctx context.Context, currentName string, newName string) (ProfilesActionResult, error) {
	w.renameProfileCalls = append(w.renameProfileCalls, [2]string{currentName, newName})
	if w.renameProfileFunc == nil {
		return ProfilesActionResult{}, nil
	}

	return w.renameProfileFunc(ctx, currentName, newName)
}

func (w *workflowStub) RemoveProfile(ctx context.Context, name string) (ProfilesActionResult, error) {
	w.removeProfileCalls = append(w.removeProfileCalls, name)
	if w.removeProfileFunc == nil {
		return ProfilesActionResult{}, nil
	}

	return w.removeProfileFunc(ctx, name)
}

func (w *workflowStub) SwitchProfile(ctx context.Context, name string) (ProfilesActionResult, error) {
	w.switchProfileCalls = append(w.switchProfileCalls, name)
	if w.switchProfileFunc == nil {
		return ProfilesActionResult{}, nil
	}

	return w.switchProfileFunc(ctx, name)
}

func (w *workflowStub) Sync(ctx context.Context, identities skill_identity.Identities) (SyncActionResult, error) {
	w.syncCalls = append(w.syncCalls, identities)
	if w.syncFunc == nil {
		return SyncActionResult{}, nil
	}

	return w.syncFunc(ctx, identities)
}

func TestModel(t *testing.T) {
	type dependencies struct {
		width    int
		height   int
		snapshot Snapshot
		workflow *workflowStub
	}

	var (
		newSkills = func(t *testing.T, sourceID string, skillCount int) []catalog.Skill {
			t.Helper()

			skills := make([]catalog.Skill, 0, skillCount)
			for index := range skillCount {
				skills = append(skills, newSkill(t, sourceID, fmt.Sprintf("skill-%02d", index+1), fmt.Sprintf("skill-%02d", index+1)))
			}

			return skills
		}
		newDefaultDependencies = func(t *testing.T) *dependencies {
			t.Helper()

			configuredSource := parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			runtime := testRuntime(t)

			return &dependencies{
				width:  120,
				height: 26,
				snapshot: buildSnapshot(
					runtime,
					source.Sources{configuredSource},
					newSkills(t, configuredSource.ID(), 6),
					newProfiles(t, "Default", mustProfile(t, "Default")),
					nil,
				),
				workflow: &workflowStub{
					addSourceFunc: func(_ context.Context, _ string) (SourceActionResult, error) {
						return SourceActionResult{}, nil
					},
					removeSourceFunc: func(_ context.Context, _ string) (SourceActionResult, error) {
						return SourceActionResult{}, nil
					},
					refreshFunc: func(_ context.Context) (RefreshActionResult, error) {
						return RefreshActionResult{}, nil
					},
					syncFunc: func(_ context.Context, _ skill_identity.Identities) (SyncActionResult, error) {
						return SyncActionResult{}, nil
					},
				},
			}
		}
		newSut = func(deps *dependencies) Model {
			sut := New(deps.snapshot, deps.workflow)
			sut.section = sectionCatalog
			if len(deps.snapshot.Sources) > 0 {
				sut.activeSourceID = deps.snapshot.Sources[0].ID()
			}
			sut.width = deps.width
			sut.height = deps.height
			sut.ready = true

			return sut
		}
		sendKey = func(t *testing.T, currentModel Model, key tea.KeyPressMsg) Model {
			t.Helper()

			updated, _ := currentModel.Update(key)
			switch updatedModel := updated.(type) {
			case Model:
				return updatedModel
			default:
				t.Fatalf("unexpected model type %T", updated)
				return Model{}
			}
		}
	)

	t.Run("init requests window size", func(t *testing.T) {
		var (
			sut = newSut(newDefaultDependencies(t))
		)

		require.NotNil(t, sut.Init())
	})

	t.Run("enter opens selected source and shows only that source skills", func(t *testing.T) {
		var (
			deps = newDefaultDependencies(t)
			sut  = newSut(deps)
		)

		sut.section = sectionSources
		sut.cursor = 0

		sut = sendKey(t, sut, tea.KeyPressMsg{Code: tea.KeyEnter})

		require.Equal(t, sectionCatalog, sut.section)
		assert.Equal(t, deps.snapshot.Sources[0].ID(), sut.activeSourceID)
		require.Len(t, sut.currentItems(), 6)
		assert.Equal(t, "skill-01", sut.currentItems()[0].Title)
	})

	t.Run("escape returns from source skills to source list", func(t *testing.T) {
		var (
			deps = newDefaultDependencies(t)
			sut  = newSut(deps)
		)

		sut.section = sectionSources
		sut = sendKey(t, sut, tea.KeyPressMsg{Code: tea.KeyEnter})
		sut = sendKey(t, sut, tea.KeyPressMsg{Code: tea.KeyEsc})

		assert.Equal(t, sectionSources, sut.section)
		assert.Empty(t, sut.activeSourceID)
		require.Len(t, sut.currentItems(), 1)
		assert.Equal(t, deps.snapshot.Sources[0].ID(), sut.currentItems()[0].Title)
	})

	t.Run("source input enter starts async add source flow", func(t *testing.T) {
		var (
			deps     = newDefaultDependencies(t)
			addedURL string
		)

		deps.workflow.addSourceFunc = func(_ context.Context, url string) (SourceActionResult, error) {
			addedURL = url
			return SourceActionResult{}, nil
		}

		sut := newSut(deps)
		sut.section = sectionSources
		sut = sendKey(t, sut, tea.KeyPressMsg{Code: 'a', Text: "a"})
		sut = sendKey(t, sut, tea.KeyPressMsg{Text: "https://github.com/anthropics/skills/tree/main/skills"})

		updated, cmd := sut.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		sut = updated.(Model)

		require.NotNil(t, cmd)
		assert.True(t, sut.refreshing)
		assert.False(t, sut.sourceInputActive)
		assert.Empty(t, addedURL)
		assert.Len(t, deps.workflow.addSourceCalls, 0)
	})

	t.Run("source added message updates snapshot and status", func(t *testing.T) {
		var (
			deps             = newDefaultDependencies(t)
			configuredSource = deps.snapshot.Sources[0]
			runtime          = deps.snapshot.Runtime
			sut              = newSut(deps)
		)

		sut.refreshing = true

		updated, _ := sut.Update(sourceAddedMsg{result: SourceActionResult{
			Snapshot: func() *Snapshot {
				snapshot := buildSnapshot(
					runtime,
					source.Sources{configuredSource},
					[]catalog.Skill{newSkill(t, configuredSource.ID(), "reviewer", "Reviewer")},
					newProfiles(t, "Default", mustProfile(t, "Default")),
					nil,
				)
				return &snapshot
			}(),
			Source:  configuredSource,
			Summary: "Added https://github.com/anthropics/skills/tree/main/skills • indexed 1 skill",
		}})
		sut = updated.(Model)

		assert.False(t, sut.refreshing)
		require.Len(t, sut.snapshot.Sources, 1)
		require.Len(t, sut.snapshot.Catalog.Skills(), 1)
		assert.Contains(t, sut.statusMessage, "Added")
	})

	t.Run("space toggles current catalog skill without syncing", func(t *testing.T) {
		var (
			deps = newDefaultDependencies(t)
			sut  = newSut(deps)
		)

		sut = sendKey(t, sut, tea.KeyPressMsg{Code: ' ', Text: " "})

		summary := sut.selectionSummary()

		assert.True(t, sut.currentItems()[0].Selected)
		assert.Equal(t, 1, summary.SelectedCount)
		assert.Equal(t, 1, summary.PendingAddCount)
		assert.Equal(t, 0, summary.PendingDelCount)
		assert.Len(t, deps.workflow.syncCalls, 0)
	})

	t.Run("s starts sync when workflow exists", func(t *testing.T) {
		var (
			deps = newDefaultDependencies(t)
			sut  = newSut(deps)
		)

		sut = sendKey(t, sut, tea.KeyPressMsg{Code: ' ', Text: " "})

		updated, cmd := sut.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
		sut = updated.(Model)

		require.NotNil(t, cmd)
		assert.True(t, sut.syncing)
		assert.Len(t, deps.workflow.syncCalls, 0)
	})

	t.Run("sync completed message updates active selection and clears dirty", func(t *testing.T) {
		var (
			deps             = newDefaultDependencies(t)
			configuredSource = deps.snapshot.Sources[0]
			identity         = newIdentity(t, configuredSource.ID(), "skill-01")
			manifest         = newManifest(t, "opencode", "/tmp/opencode", identity)
			sut              = newSut(deps)
		)

		sut = sendKey(t, sut, tea.KeyPressMsg{Code: ' ', Text: " "})
		sut.syncing = true

		updated, _ := sut.Update(syncCompletedMsg{result: SyncActionResult{
			Snapshot: func() *Snapshot {
				snapshot := buildSnapshot(
					deps.snapshot.Runtime,
					source.Sources{configuredSource},
					deps.snapshot.Catalog.Skills(),
					newProfiles(t, "Default", mustProfile(t, "Default", identity)),
					[]skillsync.Manifest{manifest},
				)
				return &snapshot
			}(),
			Result: skillsync.Result{
				DesiredCount: 1,
				Targets:      []skillsync.TargetResult{{Adapter: "opencode", RootPath: "/tmp/opencode", Linked: 1}},
			},
		}})
		sut = updated.(Model)

		assert.False(t, sut.syncing)
		assert.Equal(t, 0, sut.selectionSummary().PendingAddCount)
		assert.Equal(t, 0, sut.selectionSummary().PendingDelCount)
		assert.Contains(t, sut.statusMessage, "Synced 1 selected skill")
		require.Len(t, sut.snapshot.Manifests, 1)
	})

	t.Run("profiles section shows the active draft selection count", func(t *testing.T) {
		var (
			deps = newDefaultDependencies(t)
			sut  = newSut(deps)
		)

		sut = sendKey(t, sut, tea.KeyPressMsg{Code: ' ', Text: " "})
		sut.section = sectionProfiles
		sut.cursor = 0

		items := sut.currentItems()

		require.Len(t, items, 1)
		assert.Equal(t, "Default", items[0].Title)
		assert.Equal(t, "active • 1 selected skills", items[0].Subtitle)
	})

	t.Run("profile switch message resets the draft selection to the new active profile", func(t *testing.T) {
		var (
			deps             = newDefaultDependencies(t)
			configuredSource = deps.snapshot.Sources[0]
			identity         = newIdentity(t, configuredSource.ID(), "skill-02")
			sut              = newSut(deps)
		)

		sut = sendKey(t, sut, tea.KeyPressMsg{Code: ' ', Text: " "})
		sut.section = sectionProfiles
		sut.cursor = 1

		updated, _ := sut.Update(profileActionMsg{result: ProfilesActionResult{
			Snapshot: func() *Snapshot {
				snapshot := buildSnapshot(
					deps.snapshot.Runtime,
					source.Sources{configuredSource},
					deps.snapshot.Catalog.Skills(),
					newProfiles(t, "reviewer", mustProfile(t, "Default"), mustProfile(t, "reviewer", identity)),
					nil,
				)
				return &snapshot
			}(),
			Summary: "Switched active profile to reviewer",
		}})
		sut = updated.(Model)

		assert.Equal(t, "reviewer", sut.snapshot.Profiles.ActiveName())
		assert.Equal(t, 0, sut.selectionSummary().PendingAddCount)
		assert.Equal(t, 0, sut.selectionSummary().PendingDelCount)
		assert.Contains(t, sut.statusMessage, "Switched active profile")
	})

	t.Run("status section groups manifests by location", func(t *testing.T) {
		var (
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			identity         = newIdentity(t, configuredSource.ID(), "reviewer")
			sut              = New(buildSnapshot(
				testRuntime(t),
				source.Sources{configuredSource},
				nil,
				newProfiles(t, "Default", mustProfile(t, "Default")),
				[]skillsync.Manifest{
					newManifest(t, "ampcode", "/tmp/agents/skills", identity),
					newManifest(t, "codex", "/tmp/agents/skills", identity),
				},
			), nil)
		)

		sut.section = sectionStatus
		sut.width = 120
		sut.height = 28
		sut.ready = true

		items := sut.currentItems()

		require.Len(t, items, 2)
		assert.Equal(t, "/tmp/agents/skills", items[1].Title)
		assert.Equal(t, "ampcode, codex • 1 synced skills", items[1].Subtitle)
		assert.Contains(t, items[1].Detail, "Adapters: ampcode, codex")
		assert.Contains(t, items[1].Detail, "Synced skills: 1")
	})
}
