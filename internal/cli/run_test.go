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

type application struct {
	listSourcesResult      source.Sources
	listSourcesErr         error
	addSourceResult        source.Source
	addSourceErr           error
	removeSourceResult     source.Source
	removeSourceErr        error
	refreshCatalogResult   app.RefreshCatalogResult
	refreshCatalogErr      error
	listCatalogResult      catalog.Catalog
	listCatalogErr         error
	syncSkillIdentitiesErr error
	syncSkillIdentitiesArg skillidentity.Identities
	syncSkillIdentitiesRes skillsync.Result
	listSyncManifestsRes   []skillsync.Manifest
	listSyncManifestsErr   error
}

func (a *application) ListSources() (source.Sources, error) {
	return a.listSourcesResult, a.listSourcesErr
}

func (a *application) AddSource(locator string) (source.Sources, source.Source, error) {
	return nil, a.addSourceResult, a.addSourceErr
}

func (a *application) RemoveSource(identifier string) (source.Sources, source.Source, error) {
	return nil, a.removeSourceResult, a.removeSourceErr
}

func (a *application) RefreshCatalog(ctx context.Context) (app.RefreshCatalogResult, error) {
	return a.refreshCatalogResult, a.refreshCatalogErr
}

func (a *application) ListCatalog() (catalog.Catalog, error) {
	return a.listCatalogResult, a.listCatalogErr
}

func (a *application) SyncSkillIdentities(identities skillidentity.Identities) (skillsync.Result, error) {
	a.syncSkillIdentitiesArg = identities
	return a.syncSkillIdentitiesRes, a.syncSkillIdentitiesErr
}

func (a *application) ListSyncManifests() ([]skillsync.Manifest, error) {
	return a.listSyncManifestsRes, a.listSyncManifestsErr
}

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

		err := cli.Run([]string{"skill-switcher", "help"}, &stdout, &stdout, &application{})

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "source list")
		assert.Contains(t, stdout.String(), "sync --all")
	})

	t.Run("list configured sources", func(t *testing.T) {
		var (
			stdout           bytes.Buffer
			configuredSource = parseSource(t, "https://github.com/anthropics/skills/tree/main/skills")
			appStub          = &application{listSourcesResult: source.Sources{configuredSource}}
		)

		err := cli.Run([]string{"skill-switcher", "source", "list"}, &stdout, &stdout, appStub)

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

		appStub := &application{
			refreshCatalogResult: app.RefreshCatalogResult{
				Sources: []source.RefreshResult{{Mirror: mirror, Action: "cloned"}},
				Catalog: catalog.NewCatalog(time.Now(), newSkill(t, configuredSource.ID(), "reviewer", "Reviewer")),
			},
		}

		err = cli.Run([]string{"skill-switcher", "refresh"}, &stdout, &stdout, appStub)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "cloned")
		assert.Contains(t, stdout.String(), configuredSource.Locator())
		assert.Contains(t, stdout.String(), "1 skill")
	})

	t.Run("list catalog skills", func(t *testing.T) {
		var stdout bytes.Buffer
		appStub := &application{
			listCatalogResult: catalog.NewCatalog(time.Now(), newSkill(t, "source-a", "reviewer", "Reviewer")),
		}

		err := cli.Run([]string{"skill-switcher", "catalog", "list"}, &stdout, &stdout, appStub)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "source-a:reviewer")
		assert.Contains(t, stdout.String(), "Reviewer")
	})

	t.Run("sync explicit skill identities", func(t *testing.T) {
		var stdout bytes.Buffer
		identity := newIdentity(t, "source-a", "reviewer")
		appStub := &application{
			syncSkillIdentitiesRes: skillsync.Result{
				DesiredCount: 1,
				Targets:      []skillsync.TargetResult{{Adapter: "opencode", RootPath: "/tmp/opencode", Linked: 1}},
			},
		}

		err := cli.Run([]string{"skill-switcher", "sync", identity.Key()}, &stdout, &stdout, appStub)

		require.NoError(t, err)
		assert.Equal(t, skillidentity.NewIdentities(identity), appStub.syncSkillIdentitiesArg)
		assert.Contains(t, stdout.String(), "Synced 1 selected skill to 1 location")
		assert.Contains(t, stdout.String(), "opencode")
	})

	t.Run("sync all catalog skills", func(t *testing.T) {
		var (
			stdout   bytes.Buffer
			identity = newIdentity(t, "source-a", "reviewer")
			appStub  = &application{
				listCatalogResult: catalog.NewCatalog(time.Now(), newSkill(t, "source-a", "reviewer", "Reviewer")),
			}
		)

		err := cli.Run([]string{"skill-switcher", "sync", "--all"}, &stdout, &stdout, appStub)

		require.NoError(t, err)
		assert.Equal(t, skillidentity.NewIdentities(identity), appStub.syncSkillIdentitiesArg)
	})

	t.Run("list sync status", func(t *testing.T) {
		var stdout bytes.Buffer
		manifest, err := skillsync.NewManifest("opencode", "/tmp/opencode", newIdentity(t, "source-a", "reviewer"))
		require.NoError(t, err)

		appStub := &application{listSyncManifestsRes: []skillsync.Manifest{manifest}}

		err = cli.Run([]string{"skill-switcher", "sync", "status"}, &stdout, &stdout, appStub)

		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "opencode")
		assert.Contains(t, stdout.String(), "1 skill")
	})

	t.Run("return sync error after printing summary", func(t *testing.T) {
		var stdout bytes.Buffer
		appStub := &application{
			syncSkillIdentitiesRes: skillsync.Result{
				DesiredCount: 0,
				Targets:      []skillsync.TargetResult{{Adapter: "opencode", RootPath: "/tmp/opencode", Error: "boom"}},
			},
			syncSkillIdentitiesErr: errors.New("boom"),
		}

		err := cli.Run([]string{"skill-switcher", "sync", "clear"}, &stdout, &stdout, appStub)

		require.Error(t, err)
		assert.Contains(t, stdout.String(), "Cleared synced skills")
	})
}
