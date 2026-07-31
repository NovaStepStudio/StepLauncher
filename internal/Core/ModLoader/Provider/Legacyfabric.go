package provider

import (
	"net/http"

	"StepLauncher/internal/Core/Cache"
	"StepLauncher/internal/Core/ModLoader"
)

type LegacyFabricProvider struct {
	*AbstractFabricProvider
}

func NewLegacyFabricProvider(cacheDir string, client *http.Client, cacheMgr *cache.Manager) *LegacyFabricProvider {
	return &LegacyFabricProvider{
		AbstractFabricProvider: &AbstractFabricProvider{
			NameVal:      "legacyfabric",
			MetaBase:     "https://meta.legacyfabric.net/v2",
			MavenBase:    "https://maven.fabricmc.net/",
			CacheDir:     cacheDir,
			CacheManager: cacheMgr,
			HttpClient:   client,
		},
	}
}

var _ modloader.ModLoaderProvider = (*LegacyFabricProvider)(nil)
