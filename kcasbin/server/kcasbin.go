package kcasbin

import (
	_ "embed"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	rediswatcher "github.com/casbin/redis-watcher/v2"
	"github.com/google/uuid"
	"github.com/hkensame/goken/kcasbin/proto"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gorm.io/gorm"
)

//go:embed casbin.conf
var modelFile string

type Kasbin struct {
	Casb   *casbin.SyncedCachedEnforcer
	Logger *otelzap.Logger
	proto.UnimplementedAuthorizationServer
}

type OptionFunc func(*Kasbin)

func MustNewGormKasbin(db *gorm.DB, opts ...OptionFunc) *Kasbin {
	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, "goken", "rbac")
	if err != nil {
		panic(err)
	}
	model := model.NewModel()
	if err := model.LoadModelFromText(modelFile); err != nil {
		panic(err)
	}

	r := &Kasbin{}
	r.Casb, err = casbin.NewSyncedCachedEnforcer(model, adapter)
	if err != nil {
		panic(err)
	}

	for _, opt := range opts {
		opt(r)
	}

	if err := r.Casb.LoadPolicy(); err != nil {
		panic(err)
	}

	return r
}

func setRedisOption(opt *rediswatcher.WatcherOptions) *rediswatcher.WatcherOptions {
	opt.IgnoreSelf = true
	ud, _ := uuid.NewV7()
	opt.LocalID = ud.String()
	return opt
}

func WithRedisWatcher(addr string, opt rediswatcher.WatcherOptions) OptionFunc {
	return func(k *Kasbin) {
		opt = *setRedisOption(&opt)
		wc, err := rediswatcher.NewWatcher(addr, opt)
		if err != nil {
			panic(err)
		}
		if err := wc.SetUpdateCallback(func(s string) {}); err != nil {
			panic(err)
		}
		k.Casb.SetWatcher(wc)
	}

}

// 或许可以尝试重试
func (ks *Kasbin) LoadPolicy(str string) {
	ks.Logger.Sugar().Infof("写入触发,准备update msg = %s", str)
	if err := ks.Casb.LoadPolicy(); err != nil {
		ks.Logger.Sugar().Error("[kcasbin] LoadPolicy调用失败 err = %v", err)
	}
}

func WithLogger(l *otelzap.Logger) OptionFunc {
	return func(k *Kasbin) {
		k.Logger = l
	}
}
