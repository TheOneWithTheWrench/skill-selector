package app

import (
	"github.com/TheOneWithTheWrench/skill-selector/internal/catalog"
	"github.com/TheOneWithTheWrench/skill-selector/internal/profile"
	"github.com/TheOneWithTheWrench/skill-selector/internal/source"
	skillsync "github.com/TheOneWithTheWrench/skill-selector/internal/sync"
)

// These test contracts exist so moq can generate one mock file for the app package,
// even though App depends on interfaces declared in other packages.

type SourceRepository interface {
	source.Repository
}

type SourceRefresher interface {
	source.Refresher
}

type CatalogRepository interface {
	catalog.Repository
}

type ProfileRepository interface {
	profile.Repository
}

type SyncManifestRepository interface {
	skillsync.ManifestRepository
}

//go:generate moq -out mocks_test.go -pkg app_test . Clock SourceRepository SourceRefresher CatalogRepository ProfileRepository SyncManifestRepository
