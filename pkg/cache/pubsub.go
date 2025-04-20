package cache

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"github.com/tidwall/gjson"
)

/*
	注意后续是要引入锁的,但是因为使用发布/订阅模式时增删改是针对redis数据的,所以不会导致并发过大的问题,
*/

const (
	CacheChannel = "cache-channel"
)

const (
	VersionStr = "vrs"
	DataStr    = "data"
	KeyStr     = "key"
	ExpireStr  = "exp"
)

const batchSize = 100

const (
	//这个lua脚本用于更新更新版本的数据
	// KEYS[1] = key
	// ARGV[1] = data (JSON data)
	// ARGV[2] = vrs (number)
	// ARGV[3] = expire (seconds)
	setWithVersion = `
	local current = redis.call("GET", KEYS[1])
	if current then
		local current_version = cjson.decode(current)["vrs"]
		if tonumber(current_version) > tonumber(ARGV[2]) then
			return 0
		end
	end
	redis.call("SETEX", KEYS[1], ARGV[3], ARGV[1])
	return 1
	`

	//这个lua脚本用于删除小于等于指定版本或更小版本的数据
	// KEYS[...] = key
	// ARGV[...] = vrs
	deleteWithVersion = `
local deleted = 0

for i = 1, #KEYS do
    local current = redis.call("GET", KEYS[i])
    if current then
        local status, data = pcall(cjson.decode, current)
        if status then
            local version = tonumber(data["vrs"])
            local limit = tonumber(ARGV[i])
            if version and version <= limit then
                redis.call("DEL", KEYS[i])
                deleted = deleted + 1
            end
        end
    end
end

return deleted
`
)

type CacheUpdateMessage struct {
	Key     string `json:"key"`
	Action  string `json:"act"`
	Data    []byte `json:"data,omitempty"`
	Version int64  `json:"vrs"`
}

func (c *MultiCache) SetWithPubSub(ctx context.Context, key string, value []byte) error {
	if err := c.distributedCache.SetEx(ctx, key, value, c.ExpireTime).Err(); err != nil {
		c.Logger.Errorf("[multi-cache] 往distributed cache中设置数据失败 err= %v", err)
		return ErrBadMultiCache
	}
	msg := CacheUpdateMessage{
		Key:     key,
		Action:  "set",
		Data:    value,
		Version: 0,
	}
	b, err := jsoniter.Marshal(msg)
	if err != nil {
		c.Logger.Errorf("[multi-cache] json marshal消息失败 err=%v", err)
		return ErrMarshalFailed
	}
	c.publishHelper(ctx, CacheChannel, b)
	return nil
}

func (c *MultiCache) SetWithVersion(ctx context.Context, key string, val []byte, version int64) error {
	if !c.UseVersionControll {
		return c.SetWithPubSub(ctx, key, val)
	}

	data := CacheUpdateMessage{
		Version: version,
		Data:    val,
		Key:     key,
		Action:  "set",
	}
	b, err := jsoniter.Marshal(&data)
	if err != nil {
		c.Logger.Errorf("[multi-cache] json marshal消息失败 err=%v", err)
		return ErrMarshalFailed
	}

	if res := c.distributedCache.Eval(ctx, setWithVersion, []string{key}, b, version, c.ExpireTime); res.Err() == nil {
		affected, _ := res.Int()
		if affected == 1 {
			c.publishHelper(ctx, CacheChannel, b)
		}
	} else {
		c.Logger.Errorf("[multi-cache] set key 失败, err=%v", err)
		return ErrSetKeyFailed
	}
	return nil
}

// TODO: 后续可以在返回error里记录哪几个区间的key没被删除成功
func (c *MultiCache) DelWithPubSub(ctx context.Context, keys ...string) error {
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}
		batch := keys[i:end]
		var msg []*CacheUpdateMessage = make([]*CacheUpdateMessage, 0, end-i)

		for _, key := range batch {
			msg = append(msg, &CacheUpdateMessage{
				Key:     key,
				Action:  "delete",
				Version: 0,
			})
		}

		if err := c.distributedCache.Del(ctx, batch...).Err(); err != nil {
			c.Logger.Errorf("[multi-cache] 删除数据失败 err=%v", err)
			continue
		}

		for _, v := range msg {
			if b, err := jsoniter.Marshal(v); err == nil {
				c.publishHelper(ctx, CacheChannel, b)
			} else {
				c.Logger.Errorf("[multi-cache] json marshal消息失败 err=%v", err)
			}
		}
	}
	return nil
}

func (c *MultiCache) DelWithVersion(ctx context.Context, keys []string, vrs []int64) error {
	if !c.UseVersionControll {
		return c.DelWithPubSub(ctx, keys...)
	}
	//先用0对齐keys和vrs,如果vrs的长度大于keys倒是可以不用管
	for i := len(vrs); i < len(keys); i++ {
		vrs = append(vrs, 0)
	}

	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}
		batch := keys[i:end]
		vals := make([]interface{}, 0, end-i)
		var msg []*CacheUpdateMessage = make([]*CacheUpdateMessage, 0, end-i)

		for _, key := range batch {
			msg = append(msg, &CacheUpdateMessage{
				Key:     key,
				Action:  "delete",
				Version: 0,
			})
			vals = append(vals, 0)
		}

		if res := c.distributedCache.Eval(ctx, deleteWithVersion, batch, vals...); res.Err() == nil {
			if i, _ := res.Int(); i != len(batch) {
				c.Logger.Warnf("[multi-cache] 预定删除%dkey,实际删除%d个key", len(batch), i)
			}
		} else {
			c.Logger.Errorf("[multi-cache] 数据删除失败 err = %v", res.Err())
			continue
		}

		for _, v := range msg {
			if b, err := jsoniter.Marshal(v); err == nil {
				c.publishHelper(ctx, CacheChannel, b)
			} else {
				c.Logger.Errorf("[multi-cache] json marshal消息失败, err = %v", err)
			}
		}

	}
	return nil
}

func (c *MultiCache) SubscribeUpdate(ctx context.Context) {
	sub := c.distributedCache.Subscribe(ctx, CacheChannel)
	ch := sub.Channel()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				var update CacheUpdateMessage
				if err := jsoniter.Unmarshal([]byte(msg.Payload), &update); err != nil {
					c.Logger.Errorf("解析来自订阅channel的缓存更新消息失败, err = %v", err)
					continue
				}
				checkNew := func() bool {
					data, err := c.localCache.Get(update.Key)
					if err == nil && c.GetVersion(data) > update.Version {
						return false
					}
					return true
				}
				//防止老板本更新慢于新版本更新
				c.Mtx.Lock()

				switch update.Action {
				case "set":
					if checkNew() {
						c.localCache.Set(update.Key, update.Data)
					}

				case "delete":
					if checkNew() {
						c.localCache.Delete(update.Key)
					}
				}
				c.Mtx.Unlock()
			}
		}
	}()
}

func (c *MultiCache) publishHelper(ctx context.Context, channel string, message interface{}) {
	opt := func() (bool, error) {
		if res := c.distributedCache.Publish(ctx, channel, message); res.Err() != nil && res.Err() != redis.ErrClosed {
			return false, res.Err()
		}
		return true, nil
	}
	rt := c.newBackoffPolicy()
	if _, err := backoff.Retry(ctx, opt, backoff.WithBackOff(rt)); err != nil {
		c.Logger.Errorf("[multi-cache] 订阅发布失败, err = %v", err)
	}

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

// 与getVersion无异,只是参数换为string
func (c *MultiCache) GetVersionInString(data string) int64 {
	res := gjson.Get(data, VersionStr)
	if !res.Exists() {
		return 0
	}
	return res.Int()
}
