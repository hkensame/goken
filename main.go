package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/allegro/bigcache"
	"github.com/hkensame/goken/pkg/cache"
	"github.com/hkensame/goken/pkg/log"
	"github.com/redis/go-redis/v9"
)

type ListNode struct {
	Val  int
	Next *ListNode
}

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func BuildTree(w []string) *TreeNode {
	if len(w) == 0 {
		return nil
	}

	r := &TreeNode{}
	m := map[int]*TreeNode{}
	m[1] = r

	for i, v := range w {
		if v == "null" {
			continue
		}

		j := i + 1
		vv, ok := m[j]
		if !ok {
			fmt.Println(j)
		}
		num, _ := strconv.Atoi(v)
		vv.Val = num
		if j*2 <= len(w) {
			l := &TreeNode{}
			m[j*2] = l
			vv.Left = l
		}
		if j*2+1 <= len(w) {
			r := &TreeNode{}
			m[j*2+1] = r
			vv.Right = r
		}
	}
	return m[1]
}

type Node struct {
	Val   int
	Left  *Node
	Right *Node
	Next  *Node
}

func main() {

	ctx := context.Background()

	// 初始化 MultiCache
	cconf := redis.ClusterOptions{}
	cconf.Addrs = []string{
		"127.0.0.1:6379",
		"127.0.0.1:6380",
		"127.0.0.1:6381",
		"127.0.0.1:6382",
		"127.0.0.1:6383",
		"127.0.0.1:6384",
	}
	cconf.Password = "123"
	cconf.ReadOnly = true
	cconf.RouteByLatency = true

	bc := bigcache.DefaultConfig(10 * time.Minute)

	mc := cache.MustNewMultiCache(&cconf, &bc, cache.WithExpireTime(12*time.Minute))

	// 启动订阅协程
	mc.SubscribeUpdate(ctx)

	// 给一些缓冲时间让订阅准备好
	time.Sleep(1 * time.Second)

	key := "user1"
	val := []byte(`{"name":"Tom","age":25}`)

	err := mc.SetWithVersion(ctx, key, val, 1)
	if err != nil {
		log.Fatalf("SetWithVersion 失败: %v", err)
	}

	// 等待 pub/sub 消息传递
	time.Sleep(1 * time.Second)

	// 检查 localCache 是否生效
	data, err := mc.GetLocalCache().Get(key)
	if err != nil {
		log.Fatalf("本地缓存未命中: %v", err)
	}
	log.Infof("本地缓存数据: %s", data)

	// 写入一个版本号较小的数据，应该不会覆盖
	err = mc.SetWithVersion(ctx, key, []byte(`{"name":"Tom","age":111}`), 0)
	if err != nil {
		log.Fatalf("低版本SetWithVersion执行出错: %v", err)
	}
	// 再等一下订阅
	time.Sleep(1 * time.Second)

	// 再取本地缓存，验证版本是否被覆盖
	data, _ = mc.GetLocalCache().Get(key)
	version := mc.GetVersion(data)
	if version != 1 {
		log.Fatalf("本地缓存被低版本覆盖了，当前版本=%d", version)
	}
	fmt.Println(data)

	dt, err := mc.GetInPubSub(ctx, key)
	if err != nil {
		log.Fatalf("本地缓存获取不到要的信息")
	} else {
		fmt.Println("本地缓存获取得要的信息", string(dt))
	}

	mc.GetLocalCache().Delete(key)

	dt, err = mc.GetInPubSub(ctx, key)
	if err != nil {
		log.Fatalf("分布式缓存获取不到要的信息")
	} else {
		fmt.Println("分布式缓存获取得到要的信息", string(dt))
	}

	dt, err = mc.GetLocalCache().Get(key)
	if err != nil {
		log.Fatalf("本地缓存无法在get中被同步")
	} else {
		fmt.Println("本地缓存可以在get中被同步", string(dt))
	}

	// 测试删除（失败场景：版本过小不应删除）
	err = mc.DelWithVersion(ctx, []string{key}, []int64{99})
	if err != nil {
		log.Fatalf("DelWithVersion 执行出错: %v", err)
	}
	time.Sleep(2 * time.Second)

	_, err = mc.GetLocalCache().Get(key)
	if err != nil && err == bigcache.ErrEntryNotFound {
		log.Infof("预期缓存已经不存在,ok")
	} else {
		log.Infof("预期缓存仍存在，版本未满足删除条件: %v", err)
	}

}

func print() {
	i := 0
	maxi := 10000

	chans := make([]chan int, 5)
	for i := range chans {
		chans[i] = make(chan int)
	}
	g := sync.WaitGroup{}
	g.Add(5)

	for j := range 5 {
		m := j
		go func() {
			for {
				r, ok := <-chans[m]
				if i > maxi {
					if ok {
						for _, v := range chans {
							close(v)
						}
					}
					g.Done()
					return
				}

				fmt.Printf("协程%d: %d\n", m, r)
				i++
				chans[(m+1)%5] <- i
			}
		}()
	}
	chans[0] <- i
	g.Wait()
}

func test1() {
	const (
		goroutineCount = 5
		target         = 10000
	)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		cond          = sync.NewCond(&mu)
		current int32 = 1
	)

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				mu.Lock()
				for int(current)%goroutineCount != id {
					if int(current) > target {
						mu.Unlock()
						return
					}
					cond.Wait()
				}

				if int(current) > target {
					cond.Broadcast()
					mu.Unlock()
					return
				}

				fmt.Printf("协程%d: %d\n", id, current)
				atomic.AddInt32(&current, 1)
				cond.Broadcast()
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("所有协程完成打印")
}

func getNext(s string) []int {
	ne := make([]int, len(s))
	ne[1] = 0
	for i, j := 2, 0; i <= len(s)-1; i++ {
		for j > 0 && s[i] != s[j+1] {
			j = ne[j]
		}
		if s[i] == s[j+1] {
			j++
		}
		ne[i] = j
	}
	return ne
}

func kmp(s string, p string) int {
	fmt.Println(s, p)
	ne := getNext(p)
	fmt.Println(ne)

	for i, j := 1, 0; i <= len(s)-1; i++ {
		for j > 0 && p[j+1] != s[i] {
			j = ne[j]
		}
		if s[i] == p[j+1] {
			j++
		}
		if j == len(p)-1 {
			return j
		}
	}
	return -1
}

func strStr(haystack string, needle string) int {
	return kmp(" "+haystack, " "+needle)
}

func timekeep(f func()) time.Duration {
	t1 := time.Now()
	f()
	return time.Now().Sub(t1)
}
