package cache

import (
	"context"

	"github.com/allegro/bigcache"
)

func (c *MultiCache) getSpan(ctx context.Context) {}

func (c *MultiCache) Stats() bigcache.Stats {
	return c.localCache.Stats()
}
