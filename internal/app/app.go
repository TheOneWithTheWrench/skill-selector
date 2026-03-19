package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/agent"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/paths"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/skillidentity"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-switcher-v2/internal/sync"
)

// App coordinates core skill-switcher use cases while staying independent of any UI layer.
type App struct {
	paths             paths.Runtime
	sourceRepository  source.Repository
	sourceRefresher   source.Refresher
	catalogRepository catalog.Repository
	catalogScanner    CatalogScanner
	syncManifestRepo  skillsync.ManifestRepository
	syncTargetsLoader SyncTargetsLoader
	clock             Clock
}

// Option injects optional dependencies during App construction.
type Option func(*options) error

// CatalogScanner discovers skills inside one mirrored source subtree.
type CatalogScanner func(source.Mirror) (catalog.Skills, error)

// Clock supplies timestamps for persisted application state.
type Clock interface {
	Now() time.Time
}

// RefreshCatalogResult contains the source refresh report and the rebuilt catalog snapshot.
type RefreshCatalogResult struct {
	Sources []source.RefreshResult
	Catalog catalog.Catalog
}

// SyncTargetsLoader discovers sync targets from the current environment.
type SyncTargetsLoader func() ([]skillsync.Target, error)

type options struct {
	sourceRepository  source.Repository
	sourceRefresher   source.Refresher
	catalogRepository catalog.Repository
	catalogScanner    CatalogScanner
	syncManifestRepo  skillsync.ManifestRepository
	syncTargetsLoader SyncTargetsLoader
	clock             Clock
}

// WithSourceRepository injects custom source persistence for tests or alternate storage backends.
func WithSourceRepository(sourceRepository source.Repository) Option {
	return func(opts *options) error {
		if sourceRepository == nil {
			return fmt.Errorf("source repository required")
		}

		opts.sourceRepository = sourceRepository
		return nil
	}
}

// WithSourceRefresher injects mirror refresh behavior for tests or alternate fetch strategies.
func WithSourceRefresher(sourceRefresher source.Refresher) Option {
	return func(opts *options) error {
		if sourceRefresher == nil {
			return fmt.Errorf("source refresher required")
		}

		opts.sourceRefresher = sourceRefresher
		return nil
	}
}

// WithCatalogRepository injects catalog persistence for tests or alternate storage backends.
func WithCatalogRepository(catalogRepository catalog.Repository) Option {
	return func(opts *options) error {
		if catalogRepository == nil {
			return fmt.Errorf("catalog repository required")
		}

		opts.catalogRepository = catalogRepository
		return nil
	}
}

// WithCatalogScanner injects catalog scanning behavior for tests or alternate discovery strategies.
func WithCatalogScanner(catalogScanner CatalogScanner) Option {
	return func(opts *options) error {
		if catalogScanner == nil {
			return fmt.Errorf("catalog scanner required")
		}

		opts.catalogScanner = catalogScanner
		return nil
	}
}

// WithClock injects the time source used when persisting catalog snapshots.
func WithClock(clock Clock) Option {
	return func(opts *options) error {
		if clock == nil {
			return fmt.Errorf("clock required")
		}

		opts.clock = clock
		return nil
	}
}

// WithSyncManifestRepository injects sync manifest persistence for tests or alternate storage backends.
func WithSyncManifestRepository(syncManifestRepo skillsync.ManifestRepository) Option {
	return func(opts *options) error {
		if syncManifestRepo == nil {
			return fmt.Errorf("sync manifest repository required")
		}

		opts.syncManifestRepo = syncManifestRepo
		return nil
	}
}

// WithSyncTargetsLoader injects sync target discovery for tests or alternate environments.
func WithSyncTargetsLoader(syncTargetsLoader SyncTargetsLoader) Option {
	return func(opts *options) error {
		if syncTargetsLoader == nil {
			return fmt.Errorf("sync targets loader required")
		}

		opts.syncTargetsLoader = syncTargetsLoader
		return nil
	}
}

// New wires an App with default dependencies for any components not provided explicitly.
func New(runtime paths.Runtime, optionFuncs ...Option) (*App, error) {
	opts := options{}

	for _, optionFunc := range optionFuncs {
		if err := optionFunc(&opts); err != nil {
			return nil, err
		}
	}

	if opts.sourceRepository == nil {
		sourceRepository, err := source.NewFileRepository(runtime.SourcesFile)
		if err != nil {
			return nil, err
		}

		opts.sourceRepository = sourceRepository
	}

	if opts.sourceRefresher == nil {
		sourceRefresher, err := source.NewGitRefresher(source.ExecRunner{})
		if err != nil {
			return nil, err
		}

		opts.sourceRefresher = sourceRefresher
	}

	if opts.catalogRepository == nil {
		catalogRepository, err := catalog.NewFileRepository(runtime.CatalogFile)
		if err != nil {
			return nil, err
		}

		opts.catalogRepository = catalogRepository
	}

	if opts.catalogScanner == nil {
		opts.catalogScanner = catalog.Scan
	}

	if opts.syncManifestRepo == nil {
		syncManifestRepo, err := skillsync.NewDirectoryManifestRepository(runtime.SyncStateDir)
		if err != nil {
			return nil, err
		}

		opts.syncManifestRepo = syncManifestRepo
	}

	if opts.syncTargetsLoader == nil {
		opts.syncTargetsLoader = agent.DefaultTargets
	}

	if opts.clock == nil {
		opts.clock = realClock{}
	}

	return &App{
		paths:             runtime,
		sourceRepository:  opts.sourceRepository,
		sourceRefresher:   opts.sourceRefresher,
		catalogRepository: opts.catalogRepository,
		catalogScanner:    opts.catalogScanner,
		syncManifestRepo:  opts.syncManifestRepo,
		syncTargetsLoader: opts.syncTargetsLoader,
		clock:             opts.clock,
	}, nil
}

// ListSources returns the normalized persisted source configuration.
func (a *App) ListSources() (source.Sources, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return nil, err
	}

	return a.sourceRepository.Load()
}

// AddSource validates a supported source locator, deduplicates by source ID, and persists the next state.
func (a *App) AddSource(rawURL string) (source.Sources, source.Source, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return nil, source.Source{}, err
	}

	configuredSource, err := source.Parse(rawURL)
	if err != nil {
		return nil, source.Source{}, err
	}

	configuredSources, err := a.sourceRepository.Load()
	if err != nil {
		return nil, source.Source{}, err
	}

	nextSources, err := configuredSources.Add(configuredSource)
	if err != nil {
		return nil, source.Source{}, err
	}

	if err := a.sourceRepository.Save(nextSources); err != nil {
		return nil, source.Source{}, err
	}

	return nextSources, configuredSource, nil
}

// RemoveSource matches a source by exact URL or stable source ID and persists the next state.
func (a *App) RemoveSource(identifier string) (source.Sources, source.Source, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return nil, source.Source{}, err
	}

	configuredSources, err := a.sourceRepository.Load()
	if err != nil {
		return nil, source.Source{}, err
	}

	nextSources, removedSource, err := configuredSources.Remove(identifier)
	if err != nil {
		return nil, source.Source{}, err
	}

	if err := a.sourceRepository.Save(nextSources); err != nil {
		return nil, source.Source{}, err
	}

	return nextSources, removedSource, nil
}

// RefreshSources updates all configured mirrors and returns the successful refresh results.
func (a *App) RefreshSources(ctx context.Context) ([]source.RefreshResult, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return nil, err
	}

	mirrors, err := a.loadMirrors()
	if err != nil {
		return nil, err
	}

	results := make([]source.RefreshResult, 0, len(mirrors))
	var allErrors []error

	for _, mirror := range mirrors {
		result, err := a.sourceRefresher.Refresh(ctx, mirror)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("refresh source %q: %w", mirror.ID(), err))
			continue
		}

		results = append(results, result)
	}

	return results, errors.Join(allErrors...)
}

// RebuildCatalog rescans configured mirrors, saves the next catalog snapshot, and returns it.
func (a *App) RebuildCatalog() (catalog.Catalog, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return catalog.Catalog{}, err
	}

	mirrors, err := a.loadMirrors()
	if err != nil {
		return catalog.Catalog{}, err
	}

	var (
		allSkills catalog.Skills
		allErrors []error
	)

	for _, mirror := range mirrors {
		discoveredSkills, err := a.catalogScanner(mirror)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("scan source %q: %w", mirror.ID(), err))
			continue
		}

		allSkills = append(allSkills, discoveredSkills...)
	}

	currentCatalog := catalog.NewCatalog(a.clock.Now(), allSkills...)
	if err := a.catalogRepository.Save(currentCatalog); err != nil {
		allErrors = append(allErrors, err)
	}

	return currentCatalog, errors.Join(allErrors...)
}

// RefreshCatalog refreshes configured sources and rebuilds the catalog from the local mirrors.
func (a *App) RefreshCatalog(ctx context.Context) (RefreshCatalogResult, error) {
	refreshedSources, refreshErr := a.RefreshSources(ctx)
	currentCatalog, catalogErr := a.RebuildCatalog()

	return RefreshCatalogResult{
		Sources: refreshedSources,
		Catalog: currentCatalog,
	}, errors.Join(refreshErr, catalogErr)
}

// ListCatalog returns the current persisted catalog snapshot.
func (a *App) ListCatalog() (catalog.Catalog, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return catalog.Catalog{}, err
	}

	return a.catalogRepository.Load()
}

// ListSyncManifests returns the currently persisted sync ownership state.
func (a *App) ListSyncManifests() ([]skillsync.Manifest, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return nil, err
	}

	return a.syncManifestRepo.LoadAll()
}

// SyncSkillIdentities reconciles lightweight skill identities across detected sync targets.
func (a *App) SyncSkillIdentities(desired skillidentity.Identities) (skillsync.Result, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return skillsync.Result{}, err
	}

	targets, err := a.syncTargetsLoader()
	if err != nil {
		return skillsync.Result{}, err
	}

	manifests, err := a.syncManifestRepo.LoadAll()
	if err != nil {
		return skillsync.Result{}, err
	}

	resolver, err := a.sourceResolver()
	if err != nil {
		return skillsync.Result{}, err
	}

	result, syncErr := skillsync.Run(desired, targets, manifests, resolver)
	var allErrors []error
	if syncErr != nil {
		allErrors = append(allErrors, syncErr)
	}

	for _, manifest := range result.Manifests {
		if err := a.syncManifestRepo.Save(manifest); err != nil {
			allErrors = append(allErrors, err)
		}
	}

	return result, errors.Join(allErrors...)
}

func (a *App) loadMirrors() ([]source.Mirror, error) {
	configuredSources, err := a.sourceRepository.Load()
	if err != nil {
		return nil, err
	}

	return source.NewMirrors(configuredSources, a.paths.SourcesDir)
}

func (a *App) sourceResolver() (skillsync.Resolver, error) {
	mirrors, err := a.loadMirrors()
	if err != nil {
		return nil, err
	}

	mirrorIndex := make(map[string]source.Mirror, len(mirrors))
	for _, mirror := range mirrors {
		mirrorIndex[mirror.ID()] = mirror
	}

	return func(identity skillidentity.Identity) (string, error) {
		mirror, ok := mirrorIndex[identity.SourceID()]
		if !ok {
			return "", os.ErrNotExist
		}

		return mirror.SkillPath(identity.RelativePath()), nil
	}, nil
}

type realClock struct{}

func (realClock) Now() time.Time {
	return time.Now().UTC()
}
