package engine
import (
	"StepLauncher/internal/Core/Cache"
)

type CacheInfo = cache.Info

func (e *Engine) GetCacheInfo() CacheInfo {
	return e.cache.Info()
}

func (e *Engine) ClearAllCache() int {
	return e.cache.Clear()
}

func (e *Engine) DeleteCacheCategory(category string) int {
	return e.cache.DeleteCategory(category)
}

func (e *Engine) DeleteCacheEntry(category, key string) error {
	return e.cache.Delete(category, key)
}

func (e *Engine) RefreshCache(category, key string) error {
	return e.cache.Delete(category, key)
}
