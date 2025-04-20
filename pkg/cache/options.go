package cache

import (
	"time"

	"github.com/allegro/bigcache"
	"github.com/hkensame/goken/pkg/log"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

type OptionFunc func(*MultiCache)

func MustNewLocalCache(conf *bigcache.Config) *LocalCache {
	cache, err := bigcache.NewBigCache(*conf)
	if err != nil {
		panic(err)
	}
	return &LocalCache{
		BigCache: cache,
		conf:     conf,
	}
}

func MustNewDistributedCache(conf *redis.ClusterOptions) *DistributedCache {
	if len(conf.Addrs) <= 1 {
		panic("请直接使用单机redis")
	} else {
		cache := redis.NewClusterClient(conf)
		return &DistributedCache{
			ClusterClient: cache,
		}
	}
}

// 默认会将使用的分布式缓存作为分布式锁
func MustNewMultiCache(dc *redis.ClusterOptions, lc *bigcache.Config, opts ...OptionFunc) *MultiCache {
	c := &MultiCache{
		localCache:         MustNewLocalCache(lc),
		distributedCache:   MustNewDistributedCache(dc),
		ExpireTime:         10 * time.Minute,
		UseVersionControll: true,
		EnableTracing:      false,
	}
	for _, opt := range opts {
		opt(c)
	}

	if lc.LifeWindow >= c.ExpireTime {
		panic(ErrBadExpireTime)
	}

	if c.logger == nil {
		c.logger = log.Logger()
		c.Logger = log.Sugar()
	}

	// if c.redlock == nil {
	// 	c.redlock = redlock.MustNewRedLock(
	// 		dc.Addrs,
	// 		redlock.WithCluster(true),
	// 		redlock.WithMaxSleepTime(5*time.Second),
	// 		redlock.WithPassword(dc.Password),
	// 		redlock.WithSleepTime(10*time.Millisecond),
	// 	)
	// }
	return c
}

// func WithRedlock(r *redlock.RedLock) OptionFunc {
// 	return func(mc *MultiCache) {
// 		mc.redlock = r
// 	}
// }

func WithExpireTime(t time.Duration) OptionFunc {
	return func(mc *MultiCache) {
		mc.ExpireTime = t
	}
}

func WithVersionControl(enable bool) OptionFunc {
	return func(m *MultiCache) {
		m.UseVersionControll = enable
	}
}

// WithLogger 配置日志记录器
func WithLogger(logger *otelzap.Logger) OptionFunc {
	return func(m *MultiCache) {
		m.logger = logger
		m.Logger = logger.Sugar()
	}
}
