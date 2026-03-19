package cli_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/skill-selector/internal/app"
	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/cli"
	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-selector/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	type dependencies struct {
		Application *ApplicationMock
		OpenTUI     func() error
	}

	var (
		newDefaultDependencies = func() *dependencies {
			return &dependencies{
				Application: &ApplicationMock{
					ListSourcesFunc:         func() (source.Sources, error) { return nil, nil },
					AddSourceFunc:           func(string) (source.Sources, source.Source, error) { return nil, source.Source{}, nil },
					RemoveSourceFunc:        func(string) (source.Sources, source.Source, error) { return nil, source.Source{}, nil },
					RefreshCatalogFunc:      func(context.Context) (app.RefreshCatalogResult, error) { return app.RefreshCatalogResult{}, nil },
					ListCatalogFunc:         func() (catalog.Catalog, error) { return catalog.Catalog{}, nil },
					ListProfilesFunc:        func() (profile.Profiles, error) { return profile.DefaultProfiles(), nil },
					CreateProfileFunc:       func(string) (profile.Profiles, error) { return profile.DefaultProfiles(), nil },
					RenameProfileFunc:       func(string, string) (profile.Profiles, error) { return profile.DefaultProfiles(), nil },
					RemoveProfileFunc:       func(string) (profile.Profiles, error) { return profile.DefaultProfiles(), nil },
					SwitchProfileFunc:       func(string) (profile.Profiles, error) { return profile.DefaultProfiles(), nil },
					SyncSkillIdentitiesFunc: func(skill_identity.Identities) (skillsync.Result, error) { return skillsync.Result{}, nil },
					ListSyncManifestsFunc:   func() ([]skillsync.Manifest, error) { return nil, nil },
				},
				OpenTUI: func() error { return nil },
			}
		}
		parseSource = func(t *testing.T, rawURL string) source.Source {
			t.Helper()

			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skill_identity.Identity {
			t.Helper()

			identity, err := skill_identity.New(sourceID, relativePath)
			require.NoError(t, err)
			return identity
		}
		newSkill = func(t *testing.T, sourceID string, relativePath string, name string) catalog.Skill {
			t.Helper()

			identity := newIdentity(t, sourceID, relativePath)
			discoveredSkill, err := catalog.NewSkill(identity, name, name+" description")
			require.NoError(t, err)
			return discoveredSkill
		}
		run = func(t *testing.T, deps *dependencies, args ...string) (string, string, error) {
			t.Helper()

			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)

			err := cli.Run(append([]string{"skill-selector"}, args...), &stdout, &stderr, deps.Application, deps.OpenTUI)

			return stdout.String(), stderr.String(), err
		}
	)

	t.Run("print help when requested", func(t *testing.T) {
		var (
			deps           = newDefaultDependencies()
			stdout, _, err = run(t, deps, "help")
		)

		require.NoError(t, err)
		assert.Contains(t, stdout, "skill-selector")
		assert.Contains(t, stdout, "source list")
		assert.Contains(t, stdout, "profile")
		assert.Contains(t, stdout, "tui")
		assert.Contains(t, stdout, "sync --all")
		assert.Len(t, deps.Application.ListSourcesCalls(), 0)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListProfilesCalls(), 0)
		assert.Len(t, deps.Application.CreateProfileCalls(), 0)
		assert.Len(t, deps.Application.RenameProfileCalls(), 0)
		assert.Len(t, deps.Application.RemoveProfileCalls(), 0)
		assert.Len(t, deps.Application.SwitchProfileCalls(), 0)
		assert.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 0)
		assert.Len(t, deps.Application.ListSyncManifestsCalls(), 0)
	})

	t.Run("open tui when no command is given", func(t *testing.T) {
		var (
			openTUICalls int
			deps         = newDefaultDependencies()
		)

		deps.OpenTUI = func() error {
			openTUICalls++
			return nil
		}

		stdout, stderr, err := run(t, deps)

		require.NoError(t, err)
		assert.Empty(t, stdout)
		assert.Empty(t, stderr)
		assert.Equal(t, 1, openTUICalls)
		assert.Len(t, deps.Application.ListSourcesCalls(), 0)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListProfilesCalls(), 0)
		assert.Len(t, deps.Application.CreateProfileCalls(), 0)
		assert.Len(t, deps.Application.RenameProfileCalls(), 0)
		assert.Len(t, deps.Application.RemoveProfileCalls(), 0)
		assert.Len(t, deps.Application.SwitchProfileCalls(), 0)
		assert.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 0)
		assert.Len(t, deps.Application.ListSyncManifestsCalls(), 0)
	})

	t.Run("open tui", func(t *testing.T) {
		var (
			openTUICalls int
			deps         = newDefaultDependencies()
		)

		deps.OpenTUI = func() error {
			openTUICalls++
			return nil
		}

		_, _, err := run(t, deps, "tui")

		require.NoError(t, err)
		assert.Equal(t, 1, openTUICalls)
		assert.Len(t, deps.Application.ListSourcesCalls(), 0)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListProfilesCalls(), 0)
		assert.Len(t, deps.Application.CreateProfileCalls(), 0)
		assert.Len(t, deps.Application.RenameProfileCalls(), 0)
		assert.Len(t, deps.Application.RemoveProfileCalls(), 0)
		assert.Len(t, deps.Application.SwitchProfileCalls(), 0)
		assert.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 0)
		assert.Len(t, deps.Application.ListSyncManifestsCalls(), 0)
	})

	t.Run("list configured sources", func(t *testing.T) {
		var (
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			deps             = newDefaultDependencies()
		)

		deps.Application.ListSourcesFunc = func() (source.Sources, error) {
			return source.Sources{configuredSource}, nil
		}

		stdout, _, err := run(t, deps, "source", "list")

		require.NoError(t, err)
		assert.Contains(t, stdout, configuredSource.ID())
		assert.Contains(t, stdout, configuredSource.Locator())
		require.Len(t, deps.Application.ListSourcesCalls(), 1)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListCatalogCalls(), 0)
		assert.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 0)
		assert.Len(t, deps.Application.ListSyncManifestsCalls(), 0)
	})

	t.Run("add source", func(t *testing.T) {
		var (
			locator          = "https://github.com/anthropics/skills/tree/main/skills"
			configuredSource = parseSource(t, locator)
			deps             = newDefaultDependencies()
		)

		deps.Application.AddSourceFunc = func(gotLocator string) (source.Sources, source.Source, error) {
			return nil, configuredSource, nil
		}

		stdout, _, err := run(t, deps, "source", "add", locator)

		require.NoError(t, err)
		assert.Contains(t, stdout, "Added")
		require.Len(t, deps.Application.AddSourceCalls(), 1)
		assert.Equal(t, locator, deps.Application.AddSourceCalls()[0].S)
		assert.Len(t, deps.Application.ListSourcesCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListCatalogCalls(), 0)
		assert.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 0)
		assert.Len(t, deps.Application.ListSyncManifestsCalls(), 0)
	})

	t.Run("remove source", func(t *testing.T) {
		var (
			locator          = "https://github.com/anthropics/skills/tree/main/skills"
			configuredSource = parseSource(t, locator)
			deps             = newDefaultDependencies()
		)

		deps.Application.RemoveSourceFunc = func(identifier string) (source.Sources, source.Source, error) {
			return nil, configuredSource, nil
		}

		stdout, _, err := run(t, deps, "source", "remove", locator)

		require.NoError(t, err)
		assert.Contains(t, stdout, "Removed")
		require.Len(t, deps.Application.RemoveSourceCalls(), 1)
		assert.Equal(t, locator, deps.Application.RemoveSourceCalls()[0].S)
		assert.Len(t, deps.Application.ListSourcesCalls(), 0)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListCatalogCalls(), 0)
		assert.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 0)
		assert.Len(t, deps.Application.ListSyncManifestsCalls(), 0)
	})

	t.Run("list profiles", func(t *testing.T) {
		var (
			deps = newDefaultDependencies()
		)

		deps.Application.ListProfilesFunc = func() (profile.Profiles, error) {
			return profile.NewProfiles(
				"reviewer",
				mustProfile(t, profile.DefaultName),
				mustProfile(t, "reviewer", newIdentity(t, "source-a", "reviewer")),
			), nil
		}

		stdout, _, err := run(t, deps, "profile", "list")

		require.NoError(t, err)
		assert.Contains(t, stdout, "Default")
		assert.Contains(t, stdout, "reviewer")
		assert.Contains(t, stdout, "1 skill")
		require.Len(t, deps.Application.ListProfilesCalls(), 1)
		assert.Len(t, deps.Application.CreateProfileCalls(), 0)
		assert.Len(t, deps.Application.RenameProfileCalls(), 0)
		assert.Len(t, deps.Application.RemoveProfileCalls(), 0)
		assert.Len(t, deps.Application.SwitchProfileCalls(), 0)
	})

	t.Run("create profile", func(t *testing.T) {
		var (
			deps = newDefaultDependencies()
		)

		deps.Application.CreateProfileFunc = func(string) (profile.Profiles, error) {
			return profile.NewProfiles(profile.DefaultName, mustProfile(t, profile.DefaultName), mustProfile(t, "reviewer")), nil
		}

		stdout, _, err := run(t, deps, "profile", "create", "reviewer")

		require.NoError(t, err)
		assert.Contains(t, stdout, "Created profile reviewer")
		require.Len(t, deps.Application.CreateProfileCalls(), 1)
		assert.Equal(t, "reviewer", deps.Application.CreateProfileCalls()[0].S)
		assert.Len(t, deps.Application.ListProfilesCalls(), 0)
		assert.Len(t, deps.Application.RenameProfileCalls(), 0)
		assert.Len(t, deps.Application.RemoveProfileCalls(), 0)
		assert.Len(t, deps.Application.SwitchProfileCalls(), 0)
	})

	t.Run("rename profile", func(t *testing.T) {
		var (
			deps = newDefaultDependencies()
		)

		deps.Application.RenameProfileFunc = func(string, string) (profile.Profiles, error) {
			return profile.NewProfiles(profile.DefaultName, mustProfile(t, profile.DefaultName), mustProfile(t, "editor")), nil
		}

		stdout, _, err := run(t, deps, "profile", "rename", "reviewer", "editor")

		require.NoError(t, err)
		assert.Contains(t, stdout, "Renamed profile to editor")
		require.Len(t, deps.Application.RenameProfileCalls(), 1)
		assert.Equal(t, "reviewer", deps.Application.RenameProfileCalls()[0].S1)
		assert.Equal(t, "editor", deps.Application.RenameProfileCalls()[0].S2)
		assert.Len(t, deps.Application.ListProfilesCalls(), 0)
		assert.Len(t, deps.Application.CreateProfileCalls(), 0)
		assert.Len(t, deps.Application.RemoveProfileCalls(), 0)
		assert.Len(t, deps.Application.SwitchProfileCalls(), 0)
	})

	t.Run("remove profile", func(t *testing.T) {
		var (
			deps = newDefaultDependencies()
		)

		deps.Application.RemoveProfileFunc = func(string) (profile.Profiles, error) {
			return profile.DefaultProfiles(), nil
		}

		stdout, _, err := run(t, deps, "profile", "remove", "reviewer")

		require.NoError(t, err)
		assert.Contains(t, stdout, "Removed profile reviewer")
		require.Len(t, deps.Application.RemoveProfileCalls(), 1)
		assert.Equal(t, "reviewer", deps.Application.RemoveProfileCalls()[0].S)
		assert.Len(t, deps.Application.ListProfilesCalls(), 0)
		assert.Len(t, deps.Application.CreateProfileCalls(), 0)
		assert.Len(t, deps.Application.RenameProfileCalls(), 0)
		assert.Len(t, deps.Application.SwitchProfileCalls(), 0)
	})

	t.Run("switch profile", func(t *testing.T) {
		var (
			deps = newDefaultDependencies()
		)

		deps.Application.SwitchProfileFunc = func(string) (profile.Profiles, error) {
			return profile.NewProfiles("reviewer", mustProfile(t, profile.DefaultName), mustProfile(t, "reviewer")), nil
		}

		stdout, _, err := run(t, deps, "profile", "switch", "reviewer")

		require.NoError(t, err)
		assert.Contains(t, stdout, "Switched active profile to reviewer")
		require.Len(t, deps.Application.SwitchProfileCalls(), 1)
		assert.Equal(t, "reviewer", deps.Application.SwitchProfileCalls()[0].S)
		assert.Len(t, deps.Application.ListProfilesCalls(), 0)
		assert.Len(t, deps.Application.CreateProfileCalls(), 0)
		assert.Len(t, deps.Application.RenameProfileCalls(), 0)
		assert.Len(t, deps.Application.RemoveProfileCalls(), 0)
	})

	t.Run("refresh catalog and print source actions", func(t *testing.T) {
		var (
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			mirror, err      = source.NewMirror(configuredSource, "/tmp/sources")
			deps             = newDefaultDependencies()
		)
		require.NoError(t, err)

		deps.Application.RefreshCatalogFunc = func(context.Context) (app.RefreshCatalogResult, error) {
			return app.RefreshCatalogResult{
				Sources: []source.RefreshResult{{Mirror: mirror, Action: "cloned"}},
				Catalog: catalog.NewCatalog(time.Now(), newSkill(t, configuredSource.ID(), "reviewer", "Reviewer")),
			}, nil
		}

		stdout, _, err := run(t, deps, "refresh")

		require.NoError(t, err)
		assert.Contains(t, stdout, "cloned")
		assert.Contains(t, stdout, configuredSource.Locator())
		assert.Contains(t, stdout, "1 skill")
		require.Len(t, deps.Application.RefreshCatalogCalls(), 1)
		assert.Len(t, deps.Application.ListSourcesCalls(), 0)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.ListCatalogCalls(), 0)
		assert.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 0)
		assert.Len(t, deps.Application.ListSyncManifestsCalls(), 0)
	})

	t.Run("list catalog skills", func(t *testing.T) {
		var (
			deps = newDefaultDependencies()
		)

		deps.Application.ListCatalogFunc = func() (catalog.Catalog, error) {
			return catalog.NewCatalog(time.Now(), newSkill(t, "source-a", "reviewer", "Reviewer")), nil
		}

		stdout, _, err := run(t, deps, "catalog", "list")

		require.NoError(t, err)
		assert.Contains(t, stdout, "source-a:reviewer")
		assert.Contains(t, stdout, "Reviewer")
		require.Len(t, deps.Application.ListCatalogCalls(), 1)
		assert.Len(t, deps.Application.ListSourcesCalls(), 0)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 0)
		assert.Len(t, deps.Application.ListSyncManifestsCalls(), 0)
	})

	t.Run("sync explicit skill identities", func(t *testing.T) {
		var (
			identity = newIdentity(t, "source-a", "reviewer")
			deps     = newDefaultDependencies()
		)

		deps.Application.SyncSkillIdentitiesFunc = func(identities skill_identity.Identities) (skillsync.Result, error) {
			return skillsync.Result{
				DesiredCount: 1,
				Targets:      []skillsync.TargetResult{{Adapter: "opencode", RootPath: "/tmp/opencode", Linked: 1}},
			}, nil
		}

		stdout, _, err := run(t, deps, "sync", identity.Key())

		require.NoError(t, err)
		assert.Contains(t, stdout, "Synced 1 selected skill to 1 location")
		assert.Contains(t, stdout, "opencode")
		require.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 1)
		assert.Equal(t, skill_identity.NewIdentities(identity), deps.Application.SyncSkillIdentitiesCalls()[0].Identities)
		assert.Len(t, deps.Application.ListSourcesCalls(), 0)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListSyncManifestsCalls(), 0)
	})

	t.Run("sync all catalog skills", func(t *testing.T) {
		var (
			identity = newIdentity(t, "source-a", "reviewer")
			deps     = newDefaultDependencies()
		)

		deps.Application.ListCatalogFunc = func() (catalog.Catalog, error) {
			return catalog.NewCatalog(time.Now(), newSkill(t, "source-a", "reviewer", "Reviewer")), nil
		}
		deps.Application.SyncSkillIdentitiesFunc = func(skill_identity.Identities) (skillsync.Result, error) {
			return skillsync.Result{}, nil
		}

		_, _, err := run(t, deps, "sync", "--all")

		require.NoError(t, err)
		require.Len(t, deps.Application.ListCatalogCalls(), 1)
		require.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 1)
		assert.Equal(t, skill_identity.NewIdentities(identity), deps.Application.SyncSkillIdentitiesCalls()[0].Identities)
		assert.Len(t, deps.Application.ListSourcesCalls(), 0)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListSyncManifestsCalls(), 0)
	})

	t.Run("list sync status", func(t *testing.T) {
		var (
			manifest, err = skillsync.NewManifest("opencode", "/tmp/opencode", newIdentity(t, "source-a", "reviewer"))
			deps          = newDefaultDependencies()
		)
		require.NoError(t, err)

		deps.Application.ListSyncManifestsFunc = func() ([]skillsync.Manifest, error) {
			return []skillsync.Manifest{manifest}, nil
		}

		stdout, _, err := run(t, deps, "sync", "status")

		require.NoError(t, err)
		assert.Contains(t, stdout, "opencode")
		assert.Contains(t, stdout, "1 skill")
		require.Len(t, deps.Application.ListSyncManifestsCalls(), 1)
		assert.Len(t, deps.Application.ListSourcesCalls(), 0)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListCatalogCalls(), 0)
		assert.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 0)
	})

	t.Run("return sync error after printing summary", func(t *testing.T) {
		var (
			deps = newDefaultDependencies()
		)

		deps.Application.SyncSkillIdentitiesFunc = func(skill_identity.Identities) (skillsync.Result, error) {
			return skillsync.Result{
				DesiredCount: 0,
				Targets:      []skillsync.TargetResult{{Adapter: "opencode", RootPath: "/tmp/opencode", Error: "boom"}},
			}, errors.New("boom")
		}

		stdout, _, err := run(t, deps, "sync", "clear")

		require.Error(t, err)
		assert.Contains(t, stdout, "Cleared synced skills")
		require.Len(t, deps.Application.SyncSkillIdentitiesCalls(), 1)
		assert.Nil(t, deps.Application.SyncSkillIdentitiesCalls()[0].Identities)
		assert.Len(t, deps.Application.ListSourcesCalls(), 0)
		assert.Len(t, deps.Application.AddSourceCalls(), 0)
		assert.Len(t, deps.Application.RemoveSourceCalls(), 0)
		assert.Len(t, deps.Application.RefreshCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListCatalogCalls(), 0)
		assert.Len(t, deps.Application.ListSyncManifestsCalls(), 0)
	})
}

func mustProfile(t *testing.T, name string, identities ...skill_identity.Identity) profile.Profile {
	t.Helper()

	item, err := profile.New(name, identities...)
	require.NoError(t, err)

	return item
}
