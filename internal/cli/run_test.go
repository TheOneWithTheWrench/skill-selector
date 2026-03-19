package cli_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/app"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/cli"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	var (
		parseSource = func(t *testing.T, rawURL string) source.Source {
			t.Helper()

			configuredSource, err := source.Parse(rawURL)
			require.NoError(t, err)
			return configuredSource
		}
		newIdentity = func(t *testing.T, sourceID string, relativePath string) skillidentity.Identity {
			t.Helper()

			identity, err := skillidentity.New(sourceID, relativePath)
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
	)

	t.Run("print help when requested", func(t *testing.T) {
		var stdout bytes.Buffer

		err := cli.Run([]string{"skill-switcher", "help"}, &stdout, &stdout, &ApplicationMock{})

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "source list")
		assert.Contains(t, stdout.String(), "sync --all")
	})

	t.Run("list configured sources", func(t *testing.T) {
		var (
			stdout           bytes.Buffer
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			application      = &ApplicationMock{
				ListSourcesFunc: func() (source.Sources, error) { return source.Sources{configuredSource}, nil },
			}
		)

		err := cli.Run([]string{"skill-switcher", "source", "list"}, &stdout, &stdout, application)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), configuredSource.ID())
		assert.Contains(t, stdout.String(), configuredSource.Locator())
	})

	t.Run("refresh catalog and print source actions", func(t *testing.T) {
		var (
			stdout           bytes.Buffer
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			mirror, err      = source.NewMirror(configuredSource, "/tmp/sources")
		)
		require.NoError(t, err)

		application := &ApplicationMock{
			RefreshCatalogFunc: func(context.Context) (app.RefreshCatalogResult, error) {
				return app.RefreshCatalogResult{
					Sources: []source.RefreshResult{{Mirror: mirror, Action: "cloned"}},
					Catalog: catalog.NewCatalog(time.Now(), newSkill(t, configuredSource.ID(), "reviewer", "Reviewer")),
				}, nil
			},
		}

		err = cli.Run([]string{"skill-switcher", "refresh"}, &stdout, &stdout, application)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "cloned")
		assert.Contains(t, stdout.String(), configuredSource.Locator())
		assert.Contains(t, stdout.String(), "1 skill")
	})

	t.Run("list catalog skills", func(t *testing.T) {
		var stdout bytes.Buffer
		application := &ApplicationMock{
			ListCatalogFunc: func() (catalog.Catalog, error) {
				return catalog.NewCatalog(time.Now(), newSkill(t, "source-a", "reviewer", "Reviewer")), nil
			},
		}

		err := cli.Run([]string{"skill-switcher", "catalog", "list"}, &stdout, &stdout, application)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "source-a:reviewer")
		assert.Contains(t, stdout.String(), "Reviewer")
	})

	t.Run("sync explicit skill identities", func(t *testing.T) {
		var stdout bytes.Buffer
		identity := newIdentity(t, "source-a", "reviewer")
		application := &ApplicationMock{
			SyncSkillIdentitiesFunc: func(identities skillidentity.Identities) (skillsync.Result, error) {
				return skillsync.Result{
					DesiredCount: 1,
					Targets:      []skillsync.TargetResult{{Adapter: "opencode", RootPath: "/tmp/opencode", Linked: 1}},
				}, nil
			},
		}

		err := cli.Run([]string{"skill-switcher", "sync", identity.Key()}, &stdout, &stdout, application)

		require.NoError(t, err)
		require.Len(t, application.SyncSkillIdentitiesCalls(), 1)
		assert.Equal(t, skillidentity.NewIdentities(identity), application.SyncSkillIdentitiesCalls()[0].Identities)
		assert.Contains(t, stdout.String(), "Synced 1 selected skill to 1 location")
		assert.Contains(t, stdout.String(), "opencode")
	})

	t.Run("sync all catalog skills", func(t *testing.T) {
		var (
			stdout      bytes.Buffer
			identity    = newIdentity(t, "source-a", "reviewer")
			application = &ApplicationMock{
				ListCatalogFunc: func() (catalog.Catalog, error) {
					return catalog.NewCatalog(time.Now(), newSkill(t, "source-a", "reviewer", "Reviewer")), nil
				},
				SyncSkillIdentitiesFunc: func(identities skillidentity.Identities) (skillsync.Result, error) { return skillsync.Result{}, nil },
			}
		)

		err := cli.Run([]string{"skill-switcher", "sync", "--all"}, &stdout, &stdout, application)

		require.NoError(t, err)
		require.Len(t, application.SyncSkillIdentitiesCalls(), 1)
		assert.Equal(t, skillidentity.NewIdentities(identity), application.SyncSkillIdentitiesCalls()[0].Identities)
	})

	t.Run("list sync status", func(t *testing.T) {
		var stdout bytes.Buffer
		manifest, err := skillsync.NewManifest("opencode", "/tmp/opencode", newIdentity(t, "source-a", "reviewer"))
		require.NoError(t, err)

		application := &ApplicationMock{
			ListSyncManifestsFunc: func() ([]skillsync.Manifest, error) { return []skillsync.Manifest{manifest}, nil },
		}

		err = cli.Run([]string{"skill-switcher", "sync", "status"}, &stdout, &stdout, application)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "opencode")
		assert.Contains(t, stdout.String(), "1 skill")
	})

	t.Run("return sync error after printing summary", func(t *testing.T) {
		var stdout bytes.Buffer
		application := &ApplicationMock{
			SyncSkillIdentitiesFunc: func(skillidentity.Identities) (skillsync.Result, error) {
				return skillsync.Result{
					DesiredCount: 0,
					Targets:      []skillsync.TargetResult{{Adapter: "opencode", RootPath: "/tmp/opencode", Error: "boom"}},
				}, errors.New("boom")
			},
		}

		err := cli.Run([]string{"skill-switcher", "sync", "clear"}, &stdout, &stdout, application)

		require.Error(t, err)
		assert.Contains(t, stdout.String(), "Cleared synced skills")
	})
}
