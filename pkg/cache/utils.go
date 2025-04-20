package cache

import (
	"context"
	"strconv"
	"strings"

	"github.com/allegro/bigcache"
)

func (c *MultiCache) getSpan(ctx context.Context) {}

func (c *MultiCache) Stats() bigcache.Stats {
	return c.localCache.Stats()
}

func (c *MultiCache) joinMdData(key string, version int64, val []byte, act ...string) []byte {
	var sb strings.Builder
	sb.WriteString(`{"`)
	sb.WriteString(VersionStr)
	sb.WriteString(`":`)
	sb.WriteString(strconv.FormatInt(version, 10))

	if val != nil {
		sb.WriteString(`,"`)
		sb.WriteString(DataStr)
		sb.WriteString(`":`)
		sb.Write(val)
	}

	sb.WriteString(`,"`)
	sb.WriteString(KeyStr)
	sb.WriteString(`":"`)
	sb.WriteString(key)
	sb.WriteString(`"`)

	if len(act) > 0 {
		sb.WriteString(`,"`)
		sb.WriteString(ActionStr)
		sb.WriteString(`":"`)
		sb.WriteString(act[0])
		sb.WriteString(`"`)
	}

	sb.WriteString(`}`)
	return []byte(sb.String())
}
