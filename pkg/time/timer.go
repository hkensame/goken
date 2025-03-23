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
	closed  bool
	fn      func() bool
	doFirst bool
}

type OptionFunc func(*ResettableTimer)

// 创建新的可重置定时器,该定时器在不调用reset的情况下每隔一次指定period时间段都将运行一次fn,不会停止
// 如果调用了reset则重新进入等待而不是类似官方timer那样触发fn
func MustNewResettableTimer(period time.Duration, fn func() bool, opts ...OptionFunc) *ResettableTimer {
	rt := &ResettableTimer{
		period: period,
		fn:     fn,
		closed: false,
	}
	rt.timer = time.NewTimer(period)
	return rt
}

func (rf *ResettableTimer) Run() {
	go rf.runLoop()
}

// 运行定时器的主循环
func (rt *ResettableTimer) runLoop() {
	rt.mu.Lock()
	if !rt.closed && rt.doFirst {
		if !rt.fn() {
			rt.closed = true
		}
	}
	rt.mu.Unlock()

	for rt.closed {
		<-rt.timer.C
		rt.mu.Lock()
		if !rt.closed {
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
	rt.ResetWithTTK(rt.period)
}

func (rt *ResettableTimer) ResetWithTTK(t time.Duration) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	if !rt.closed {
		return
	}

	if !rt.timer.Stop() {
		select {
		case <-rt.timer.C: // 清空过期的定时器信号
		default:
		}
	}
	rt.timer.Reset(t)
}

// 停止定时器
func (rt *ResettableTimer) Stop() {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.closed = true
	rt.timer.Stop()
}

// 让整个定时器先执行一次传入的func再进行定时
func (rt *ResettableTimer) WithDoFirst(do bool) OptionFunc {
	return func(rt *ResettableTimer) {
		rt.doFirst = do
	}
}
