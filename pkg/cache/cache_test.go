package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gihub.com/hkensame/goken/pkg/cache" // 替换为你的包名

	"github.com/allegro/bigcache/v3"
	"github.com/redis/go-redis/v9"
)

type stdLogger struct{}

func (stdLogger) Errorf(format string, args ...interface{}) {
	fmt.Printf("[ERROR] "+format+"\n", args...)
}
func (stdLogger) Warnf(format string, args ...interface{}) {
	fmt.Printf("[WARN] "+format+"\n", args...)
}

func newLocalCache() cache.LocalCache {
	bc, _ := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	return &cache.BigCacheWrapper{Cache: bc}
}

func TestMultiCache_SetGetDelWithVersion(t *testing.T) {
	ctx := context.Background()

	// 初始化 MultiCache
	mc := &cache.MultiCache{
		distributedCache: redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: []string{
				"192.168.199.128:6379",
				"192.168.199.128:6380",
				"192.168.199.128:6381",
				"192.168.199.128:6382",
				"192.168.199.128:6383",
				"192.168.199.128:6384",
			},
			Password: "123",
		}),
		localCache:         newLocalCache(),
		Logger:             stdLogger{},
		ExpireTime:         30 * time.Second,
		UseVersionControll: true,
	}

	// 启动订阅协程
	mc.SubscribeUpdate(ctx)

	// 给一些缓冲时间让订阅准备好
	time.Sleep(1 * time.Second)

	// 测试写入
	key := "user:1001"
	val := []byte(`{"name":"Tom","age":25,"vrs":1}`)
	err := mc.SetWithVersion(ctx, key, val, 1)
	if err != nil {
		t.Fatalf("SetWithVersion 失败: %v", err)
	}

	// 等待 pub/sub 消息传递
	time.Sleep(1 * time.Second)

	// 检查 localCache 是否生效
	data, err := mc.localCache.Get(key)
	if err != nil {
		t.Fatalf("本地缓存未命中: %v", err)
	}
	t.Logf("本地缓存数据: %s", data)

	// 写入一个版本号较小的数据，应该不会覆盖
	err = mc.SetWithVersion(ctx, key, []byte(`{"name":"Jerry","vrs":0}`), 0)
	if err != nil {
		t.Fatalf("低版本SetWithVersion执行出错: %v", err)
	}
	// 再等一下订阅
	time.Sleep(1 * time.Second)

	// 再取本地缓存，验证版本是否被覆盖
	data, _ = mc.localCache.Get(key)
	version := mc.GetVersion(data)
	if version != 1 {
		t.Fatalf("本地缓存被低版本覆盖了，当前版本=%d", version)
	}

	// 测试删除（失败场景：版本过小不应删除）
	err = mc.DelWithVersion(ctx, []string{key}, []int64{0})
	if err != nil {
		t.Fatalf("DelWithVersion 执行出错: %v", err)
	}
	time.Sleep(1 * time.Second)

	_, err = mc.localCache.Get(key)
	if err != nil {
		t.Logf("预期缓存仍存在，版本未满足删除条件: %v", err)
	}

	// 正确版本删除
	err = mc.DelWithVersion(ctx, []string{key}, []int64{2})
	if err != nil {
		t.Fatalf("DelWithVersion 执行出错: %v", err)
	}
	time.Sleep(1 * time.Second)

	_, err = mc.localCache.Get(key)
	if err == nil {
		t.Fatalf("缓存未被删除（版本足够）")
	}
}
