package time

import (
	"sync"
	"time"
)

// 可重置定时器结构体
type ResettableTimer struct {
	mu      sync.Mutex
	timer   *time.Timer
	period  time.Duration
	running bool
	fn      func()
}

// 创建新的可重置定时器,该定时器在不调用reset的情况下每隔一次指定period时间段都将运行一次fn,不会停止
// 如果调用了reset则重新进入等待而不是类似官方timer那样触发fn
func MustNewResettableTimer(period time.Duration, fn func()) *ResettableTimer {
	rt := &ResettableTimer{
		period:  period,
		fn:      fn,
		running: true,
	}
	rt.timer = time.NewTimer(period)

	go rt.runLoop()
	return rt
}

// 运行定时器的主循环
func (rt *ResettableTimer) runLoop() {
	for rt.running {
		<-rt.timer.C
		rt.mu.Lock()
		if !rt.running {
			rt.mu.Unlock()
			return
		}
		rt.fn()
		rt.timer.Reset(rt.period)
		rt.mu.Unlock()
	}
}

// 重新开始定时器
func (rt *ResettableTimer) Reset() {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if !rt.running {
		return
	}

	if !rt.timer.Stop() {
		select {
		case <-rt.timer.C: // 清空过期的定时器信号
		default:
		}
	}
	rt.timer.Reset(rt.period)
}

// 停止定时器
func (rt *ResettableTimer) Stop() {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.running = false
	rt.timer.Stop()
}
