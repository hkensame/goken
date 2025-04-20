package cache

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/allegro/bigcache"
	"github.com/cenkalti/backoff/v5"
	"github.com/tidwall/gjson"
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

	if key != "" {
		sb.WriteString(`,"`)
		sb.WriteString(KeyStr)
		sb.WriteString(`":"`)
		sb.WriteString(key)
		sb.WriteString(`"`)
	}

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

func (c *MultiCache) newBackoffPolicy() *backoff.ExponentialBackOff {
	policy := backoff.NewExponentialBackOff()
	policy.InitialInterval = 1 * time.Second
	policy.MaxInterval = 16 * time.Second
	policy.Multiplier = 2
	return policy
}

// 默认情况下如果没有version字段则默认认为字段version为0
func (c *MultiCache) GetVersion(data []byte) int64 {
	res := gjson.GetBytes(data, VersionStr)
	if !res.Exists() {
		return 0
	}
	return res.Int()
}

func (c *MultiCache) GetData(data []byte) []byte {
	res := gjson.GetBytes(data, DataStr)
	if !res.Exists() {
		return nil
	}
	return []byte(res.String())
}

// 与getVersion无异,只是参数换为string
func (c *MultiCache) GetVersionInString(data string) int64 {
	res := gjson.Get(data, VersionStr)
	if !res.Exists() {
		return 0
	}
	return res.Int()
}
