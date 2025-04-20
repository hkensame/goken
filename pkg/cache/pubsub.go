package cache

import (
	"bytes"
	"context"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/cenkalti/backoff/v5"
	"github.com/hkensame/goken/pkg/log"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"github.com/tidwall/sjson"
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
	ActionStr  = "act"
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
	if err := c.distributedCache.SetEx(ctx, key, c.joinMdData("", 0, value), c.ExpireTime).Err(); err != nil {
		c.Logger.Errorf("[multi-cache] 往distributed cache中设置数据失败 err = %v", err)
		return ErrBadMultiCache
	}
	b := c.joinMdData(key, 0, value, "set")
	c.publishHelper(ctx, CacheChannel, b)
	return nil
}

func (c *MultiCache) SetWithVersion(ctx context.Context, key string, val []byte, version int64) error {
	if !c.UseVersionControll {
		return c.SetWithPubSub(ctx, key, val)
	}
	res := c.distributedCache.Eval(ctx, setWithVersion, []string{key}, c.joinMdData("", version, val), version, c.ExpireTime)
	if res.Err() == nil {
		affected, _ := res.Int()
		if affected == 1 {
			b := c.joinMdData(key, version, val, "set")
			c.publishHelper(ctx, CacheChannel, b)
		}
	} else {
		c.Logger.Errorf("[multi-cache] set key 失败, err = %v", res.Err())
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
		var msg [][]byte = make([][]byte, 0, end-i)

		for _, key := range batch {
			msg = append(msg, c.joinMdData(key, 0, nil, "delete"))
		}

		if err := c.distributedCache.Del(ctx, batch...).Err(); err != nil {
			c.Logger.Errorf("[multi-cache] 删除数据失败 err=%v", err)
			continue
		}

		for _, v := range msg {
			c.publishHelper(ctx, CacheChannel, v)

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
		vrss := make([]interface{}, 0, end-i)
		var msg [][]byte = make([][]byte, 0, end-i)

		for i, key := range batch {
			msg = append(msg, c.joinMdData(key, vrs[i], nil, "delete"))
			vrss = append(vrss, vrs[i])
		}

		if res := c.distributedCache.Eval(ctx, deleteWithVersion, batch, vrss...); res.Err() == nil {
			if i, _ := res.Int(); i != len(batch) {
				c.Logger.Warnf("[multi-cache] 预定删除%dkey,实际删除%d个key", len(batch), i)
			}
		} else {
			c.Logger.Errorf("[multi-cache] 数据删除失败 err = %v", res.Err())
			continue
		}

		for _, v := range msg {
			c.publishHelper(ctx, CacheChannel, v)
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
				td := []byte(msg.Payload)
				if err := jsoniter.Unmarshal(td, &update); err != nil {
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
						//删去要存储的额外信息
						//之前的unmarshal成功则一定保证这里删除是成功的
						td, _ = sjson.DeleteBytes(td, ActionStr)
						td, _ = sjson.DeleteBytes(td, KeyStr)
						c.localCache.Set(update.Key, td)
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

func (c *MultiCache) GetInPubSub(ctx context.Context, key string) ([]byte, error) {
	c.Mtx.RLock()
	data, err := c.localCache.Get(key)
	c.Mtx.RUnlock()
	if err != nil {
		log.Warnf("[multi-cache] %s未命中本地缓存", key)
	} else {
		return c.GetData(data), nil
	}

	res := c.distributedCache.Get(ctx, key)
	if err := res.Err(); err != nil {
		if err == redis.Nil {
			log.Warnf("[multi-cache] %s未命中分布式缓存", key)
			return nil, ErrRecordNotFound
		}
		log.Errorf("[multi-cache] 从distributed cache中获取数据失败 err = %v", err)
		return nil, ErrBadMultiCache
	}

	//因为整个multiCache都只支持存[]byte,故不用担心这里出错
	data, _ = res.Bytes()

	//因为在上面的过程中可能存在有其他携程写入了localCache,所以需要再比较一次
	c.Mtx.Lock()
	nd, err := c.localCache.Get(key)
	if err == nil && c.GetVersion(nd) > c.GetVersion(data) {
		c.Mtx.Unlock()
		return c.GetData(nd), nil
	} else {
		c.localCache.Set(key, data)
		c.Mtx.Unlock()
		return data, nil
	}
}

// 逻辑跟joinMdData一样
func (c CacheUpdateMessage) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(`{"`)
	buf.WriteString("vrs")
	buf.WriteString(`":`)
	buf.WriteString(strconv.FormatInt(c.Version, 10))

	if len(c.Data) > 0 {
		buf.WriteString(`,"data":`)
		buf.Write(c.Data)
	}

	buf.WriteString(`,"key":"`)
	buf.WriteString(c.Key)
	buf.WriteString(`","act":"`)
	buf.WriteString(c.Action)
	buf.WriteString(`"}`)
	return buf.Bytes(), nil
}

func (c *CacheUpdateMessage) UnmarshalJSON(data []byte) error {
	var err error
	if c.Version, err = jsonparser.GetInt(data, "vrs"); err != nil {
		return ErrUnmarshalFailed
	}

	if c.Key, err = jsonparser.GetString(data, "key"); err != nil {
		return ErrUnmarshalFailed
	}

	if c.Action, err = jsonparser.GetString(data, "act"); err != nil {
		return ErrUnmarshalFailed
	}

	if val, t, _, err := jsonparser.Get(data, "data"); err == nil && t != jsonparser.NotExist {
		c.Data = make([]byte, len(val))
		copy(c.Data, val)
	} else {
		c.Data = nil
	}
	return nil
}
