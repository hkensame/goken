package authmodel

/*
	注意暂时只写了一个比较简单的框架,没有往里面加入链路追踪等信息
*/

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/hkensame/goken/kauth/gserver"
	"github.com/hkensame/goken/pkg/errors"
	"github.com/hkensame/goken/server/httpserver"
	"github.com/hkensame/goken/server/httpserver/middlewares/jwt"
	rediscache "github.com/hkensame/redis"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gorm.io/gorm"
)

var (
	ErrFailedAuthentication = errors.New("用户认证失败")
)

type Auther struct {
	Cache         *rediscache.TokenStore
	Manager       *manage.Manager
	OServer       *gserver.Server
	Server        *httpserver.Server
	Jwt           *jwt.GinJWTMiddleware
	DB            *GormClientStore
	Authenticator func(*gin.Context) bool
	Coder         *CodeGener
	Logger        *otelzap.Logger
}

type OptionFunc func(*Auther)

func MustNewRedisAuther(host string, rdb *redis.Client, jwt *jwt.GinJWTMiddleware, opts ...OptionFunc) *Auther {
	r := &Auther{}
	r.Cache = rediscache.NewRedisStoreWithCli(rdb)
	r.Jwt = jwt
	r.Jwt.UseAbort = true

	for _, opt := range opts {
		opt(r)
	}

	r.Manager = manage.NewManager()
	r.Manager.MapTokenStorage(r.Cache)
	r.Manager.MapAccessGenerate(r)

	if r.Coder != nil {
		r.Manager.MapAuthorizeGenerate(r.Coder)
	}
	cfg := &manage.Config{}
	cfg.IsGenerateRefresh = true
	cfg.RefreshTokenExp = r.Jwt.MaxRefreshFunc(nil)
	cfg.AccessTokenExp = r.Jwt.TimeoutFunc(nil)
	r.Manager.SetAuthorizeCodeTokenCfg(cfg)
	r.Manager.SetClientTokenCfg(cfg)
	//r.Manager.SetRefreshTokenCfg()
	if r.DB != nil {
		r.Manager.MapClientStorage(r.DB)
	}

	conf := gserver.NewConfig()

	r.OServer = gserver.MustNewServer(conf, r.Manager, rdb)

	r.Server = httpserver.MustNewServer(context.Background(), host)

	if r.Authenticator == nil {
		panic("认证函数不能为nil")
	}
	return r
}

func WithDB(db *gorm.DB) OptionFunc {
	return func(a *Auther) {
		a.DB = MustNewGormClientStore(db)
	}
}

func WithAuthenticator(auth func(*gin.Context) bool) OptionFunc {
	return func(a *Auther) {
		a.Authenticator = auth
	}
}

func WithCodeGener(c *CodeGener) OptionFunc {
	return func(a *Auther) {
		a.Coder = c
	}
}

func WithLogger(l *otelzap.Logger) OptionFunc {
	return func(a *Auther) {
		a.Logger = l
	}
}

func (r *Auther) Serve() error {
	return r.Server.Serve()
}
