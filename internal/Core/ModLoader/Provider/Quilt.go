package provider

import (
	"net/http"

	"StepLauncher/internal/Core/Cache"
	"StepLauncher/internal/Core/ModLoader"
)

type QuiltProvider struct {
	*AbstractFabricProvider
}

func NewQuiltProvider(cacheDir string, client *http.Client, cacheMgr *cache.Manager) *QuiltProvider {
	return &QuiltProvider{
		AbstractFabricProvider: &AbstractFabricProvider{
			NameVal:      "quilt",
			MetaBase:     "https://meta.quiltmc.org/v3",
			MavenBase:    "https://maven.fabricmc.net/",
			CacheDir:     cacheDir,
			CacheManager: cacheMgr,
			HttpClient:   client,
		},
	}
}

var _ modloader.ModLoaderProvider = (*QuiltProvider)(nil)
