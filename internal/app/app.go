package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/TheOneWithTheWrench/skill-selector/internal/agent"
	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/paths"
	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/skill_identity"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-selector/internal/sync"
)

// App coordinates core skill-selector use cases while staying independent of any UI layer.
type App struct {
	paths             paths.Runtime
	sourceRepository  source.Repository
	sourceRefresher   source.Refresher
	catalogRepository catalog.Repository
	profileRepository profile.Repository
	catalogScanner    CatalogScanner
	syncManifestRepo  skillsync.ManifestRepository
	syncTargetsLoader SyncTargetsLoader
	clock             Clock
}

// Option injects optional dependencies during App construction.
type Option func(*options)

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
	profileRepository profile.Repository
	catalogScanner    CatalogScanner
	syncManifestRepo  skillsync.ManifestRepository
	syncTargetsLoader SyncTargetsLoader
	clock             Clock
}

// WithSourceRepository injects custom source persistence for tests or alternate storage backends.
func WithSourceRepository(sourceRepository source.Repository) Option {
	return func(opts *options) {
		opts.sourceRepository = sourceRepository
	}
}

// WithSourceRefresher injects mirror refresh behavior for tests or alternate fetch strategies.
func WithSourceRefresher(sourceRefresher source.Refresher) Option {
	return func(opts *options) {
		opts.sourceRefresher = sourceRefresher
	}
}

// WithCatalogRepository injects catalog persistence for tests or alternate storage backends.
func WithCatalogRepository(catalogRepository catalog.Repository) Option {
	return func(opts *options) {
		opts.catalogRepository = catalogRepository
	}
}

// WithProfileRepository injects profile persistence for tests or alternate storage backends.
func WithProfileRepository(profileRepository profile.Repository) Option {
	return func(opts *options) {
		opts.profileRepository = profileRepository
	}
}

// WithCatalogScanner injects catalog scanning behavior for tests or alternate discovery strategies.
func WithCatalogScanner(catalogScanner CatalogScanner) Option {
	return func(opts *options) {
		opts.catalogScanner = catalogScanner
	}
}

// WithClock injects the time source used when persisting catalog snapshots.
func WithClock(clock Clock) Option {
	return func(opts *options) {
		opts.clock = clock
	}
}

// WithSyncManifestRepository injects sync manifest persistence for tests or alternate storage backends.
func WithSyncManifestRepository(syncManifestRepo skillsync.ManifestRepository) Option {
	return func(opts *options) {
		opts.syncManifestRepo = syncManifestRepo
	}
}

// WithSyncTargetsLoader injects sync target discovery for tests or alternate environments.
func WithSyncTargetsLoader(syncTargetsLoader SyncTargetsLoader) Option {
	return func(opts *options) {
		opts.syncTargetsLoader = syncTargetsLoader
	}
}

// New wires an App with default dependencies for any components not provided explicitly.
func New(runtime paths.Runtime, optionFuncs ...Option) (*App, error) {
	opts, err := newDefaultOptions(runtime)
	if err != nil {
		return nil, err
	}

	for _, optionFunc := range optionFuncs {
		optionFunc(&opts)
	}

	return &App{
		paths:             runtime,
		sourceRepository:  opts.sourceRepository,
		sourceRefresher:   opts.sourceRefresher,
		catalogRepository: opts.catalogRepository,
		profileRepository: opts.profileRepository,
		catalogScanner:    opts.catalogScanner,
		syncManifestRepo:  opts.syncManifestRepo,
		syncTargetsLoader: opts.syncTargetsLoader,
		clock:             opts.clock,
	}, nil
}

func newDefaultOptions(runtime paths.Runtime) (options, error) {
	sourceRepository, err := source.NewFileRepository(runtime.SourcesFile)
	if err != nil {
		return options{}, err
	}

	sourceRefresher, err := source.NewGitRefresher(source.ExecRunner{})
	if err != nil {
		return options{}, err
	}

	catalogRepository, err := catalog.NewFileRepository(runtime.CatalogFile)
	if err != nil {
		return options{}, err
	}

	profileRepository, err := profile.NewFileRepository(runtime.ProfilesFile)
	if err != nil {
		return options{}, err
	}

	syncManifestRepo, err := skillsync.NewDirectoryManifestRepository(runtime.SyncStateDir)
	if err != nil {
		return options{}, err
	}

	return options{
		sourceRepository:  sourceRepository,
		sourceRefresher:   sourceRefresher,
		catalogRepository: catalogRepository,
		profileRepository: profileRepository,
		catalogScanner:    catalog.Scan,
		syncManifestRepo:  syncManifestRepo,
		syncTargetsLoader: agent.DefaultTargets,
		clock:             realClock{},
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

	profiles, err := a.profileRepository.Load()
	if err != nil {
		return nil, source.Source{}, err
	}

	nextSources, removedSource, err := configuredSources.Remove(identifier)
	if err != nil {
		return nil, source.Source{}, err
	}
	nextProfiles := profiles.WithoutSource(removedSource.ID())

	if err := a.sourceRepository.Save(nextSources); err != nil {
		return nil, source.Source{}, err
	}
	if err := a.profileRepository.Save(nextProfiles); err != nil {
		rollbackErr := a.sourceRepository.Save(configuredSources)
		if rollbackErr != nil {
			return nil, source.Source{}, errors.Join(err, fmt.Errorf("rollback removed source %q: %w", removedSource.ID(), rollbackErr))
		}

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

// ListProfiles returns the currently persisted profile state.
func (a *App) ListProfiles() (profile.Profiles, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return profile.Profiles{}, err
	}

	return a.profileRepository.Load()
}

// CreateProfile adds a new empty profile and persists the next profile state.
func (a *App) CreateProfile(name string) (profile.Profiles, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return profile.Profiles{}, err
	}

	profiles, err := a.profileRepository.Load()
	if err != nil {
		return profile.Profiles{}, err
	}

	nextProfiles, err := profiles.Create(name)
	if err != nil {
		return profile.Profiles{}, err
	}

	if err := a.profileRepository.Save(nextProfiles); err != nil {
		return profile.Profiles{}, err
	}

	return nextProfiles, nil
}

// RenameProfile renames one stored profile and persists the next profile state.
func (a *App) RenameProfile(currentName string, newName string) (profile.Profiles, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return profile.Profiles{}, err
	}

	profiles, err := a.profileRepository.Load()
	if err != nil {
		return profile.Profiles{}, err
	}

	nextProfiles, err := profiles.Rename(currentName, newName)
	if err != nil {
		return profile.Profiles{}, err
	}

	if err := a.profileRepository.Save(nextProfiles); err != nil {
		return profile.Profiles{}, err
	}

	return nextProfiles, nil
}

// RemoveProfile removes one inactive profile and persists the next profile state.
func (a *App) RemoveProfile(name string) (profile.Profiles, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return profile.Profiles{}, err
	}

	profiles, err := a.profileRepository.Load()
	if err != nil {
		return profile.Profiles{}, err
	}

	nextProfiles, err := profiles.Remove(name)
	if err != nil {
		return profile.Profiles{}, err
	}

	if err := a.profileRepository.Save(nextProfiles); err != nil {
		return profile.Profiles{}, err
	}

	return nextProfiles, nil
}

// SwitchProfile changes the active profile without syncing it automatically.
func (a *App) SwitchProfile(name string) (profile.Profiles, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return profile.Profiles{}, err
	}

	profiles, err := a.profileRepository.Load()
	if err != nil {
		return profile.Profiles{}, err
	}

	nextProfiles, err := profiles.Switch(name)
	if err != nil {
		return profile.Profiles{}, err
	}

	if err := a.profileRepository.Save(nextProfiles); err != nil {
		return profile.Profiles{}, err
	}

	return nextProfiles, nil
}

// SaveActiveProfileSelection persists the saved selection owned by the active profile.
func (a *App) SaveActiveProfileSelection(desired skill_identity.Identities) (profile.Profiles, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return profile.Profiles{}, err
	}

	profiles, err := a.profileRepository.Load()
	if err != nil {
		return profile.Profiles{}, err
	}

	nextProfiles, err := profiles.SetActiveSelection(desired)
	if err != nil {
		return profile.Profiles{}, err
	}

	if err := a.profileRepository.Save(nextProfiles); err != nil {
		return profile.Profiles{}, err
	}

	return nextProfiles, nil
}

// ListSyncManifests returns the currently persisted sync ownership state.
func (a *App) ListSyncManifests() ([]skillsync.Manifest, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return nil, err
	}

	return a.syncManifestRepo.LoadAll()
}

// SyncSkillIdentities reconciles lightweight skill identities across detected sync targets.
func (a *App) SyncSkillIdentities(desired skill_identity.Identities) (skillsync.Result, error) {
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

	return func(identity skill_identity.Identity) (string, error) {
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
