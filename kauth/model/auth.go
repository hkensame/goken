package authmodel

/*
	注意暂时只写了一个比较简单的框架,没有往里面加入链路追踪等信息
*/

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
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
	OServer       *server.Server
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

	if r.DB != nil {
		r.Manager.MapClientStorage(r.DB)
	}

	//AllowedGrantTypes说明
	//AuthorizationCode	拿到了 code 后要换 token
	//Password	传统用户登录(账号+密码) 现在几乎不使用
	//ClientCredentials	机器对机器 / 服务对服务调用
	//RefreshToken	使用旧 token 获取新 token
	conf := server.NewConfig()
	conf.AllowedResponseTypes = []oauth2.ResponseType{oauth2.Code}
	conf.AllowedGrantTypes = []oauth2.GrantType{
		oauth2.AuthorizationCode,
		//oauth2.PasswordCredentials,
		oauth2.ClientCredentials,
		//oauth2.Refreshing,
	}
	conf.AllowGetAccessRequest = true
	conf.AllowedCodeChallengeMethods = []oauth2.CodeChallengeMethod{oauth2.CodeChallengeS256}
	conf.ForcePKCE = true
	r.OServer = server.NewServer(conf, r.Manager)

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
