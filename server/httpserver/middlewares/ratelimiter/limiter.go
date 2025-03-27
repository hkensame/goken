package ratelimiter

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hkensame/goken/pkg/common/httputil"
	tokenbuc "github.com/juju/ratelimit"
	leakybuc "go.uber.org/ratelimit"
)

type FailedBehavior func(ctx *gin.Context, l *tokenbuc.Bucket)

var (
	MaxWaitingTime time.Duration = 3 * time.Second
)

var (
	//直接abort
	Abort FailedBehavior = func(ctx *gin.Context, l *tokenbuc.Bucket) {
		//未获取成功就退出
		if l.TakeAvailable(1) <= 0 {
			httputil.WriteResponse(ctx, 500, "服务器繁忙,请稍后再试", nil, true)
			return
		}
		ctx.Next()
	}

	Wait FailedBehavior = func(ctx *gin.Context, l *tokenbuc.Bucket) {
		if !l.WaitMaxDuration(1, MaxWaitingTime) {
			httputil.WriteResponse(ctx, 500, "服务器繁忙,请稍后再试", nil, true)
			return
		}
		ctx.Next()
	}

	//允许自定义等待时间
	WaitWithTime func(d time.Duration) FailedBehavior = func(d time.Duration) FailedBehavior {
		return func(ctx *gin.Context, l *tokenbuc.Bucket) {
			if !l.WaitMaxDuration(1, d) {
				httputil.WriteResponse(ctx, 500, "服务器繁忙,请稍后再试", nil, true)
				return
			}
			ctx.Next()
		}
	}
)

// 若不通过option修改则默认每秒允许rate个流量
func LeakyBucketHandler(rat int, opts ...leakybuc.Option) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limiter := leakybuc.New(rat, opts...)
		limiter.Take()
		ctx.Next()
	}
}

func TokenBucketHandler(interval time.Duration, cap int64, behavior FailedBehavior) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limiter := tokenbuc.NewBucket(interval, cap)
		behavior(ctx, limiter)
	}
}
