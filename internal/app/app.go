package app

import (
	"fmt"

	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/paths"
	"github.com/TheOneWithTheWrench/skill-switcher-v2/internal/source"
)

// App coordinates core skill-switcher use cases while staying independent of any UI layer.
type App struct {
	paths            paths.Runtime
	sourceRepository source.Repository
}

// Option injects optional dependencies during App construction.
type Option func(*options) error

type options struct {
	sourceRepository source.Repository
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

	return &App{
		paths:            runtime,
		sourceRepository: opts.sourceRepository,
	}, nil
}

// ListSources returns the normalized persisted source configuration.
func (a *App) ListSources() (source.Sources, error) {
	if err := a.paths.EnsureRuntimeDirs(); err != nil {
		return nil, err
	}

	return a.sourceRepository.Load()
}

// AddSource validates a GitHub tree URL, deduplicates by source ID, and persists the next state.
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
