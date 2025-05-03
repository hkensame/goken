package authmodel

/*
	注意暂时只写了一个比较简单的框架,没有往里面加入链路追踪等信息
*/

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hkensame/goken/kauth/gserver"
	"github.com/hkensame/goken/pkg/errors"
	"github.com/hkensame/goken/server/httpserver"
	"github.com/hkensame/goken/server/httpserver/middlewares/jwt"
	rediscache "github.com/hkensame/redis"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"gorm.io/gorm"
)

var (
	ErrFailedAuthentication = errors.New("用户认证失败")
)

type Auther struct {
	Cache         *rediscache.TokenStore
	OServer       *gserver.Server
	Server        *httpserver.Server
	Jwt           *jwt.GinJWTMiddleware
	DB            *GormClientStore
	Authenticator func(*gin.Context) bool
	logger        *otelzap.Logger
	Logger        *otelzap.SugaredLogger
}

type OptionFunc func(*Auther)

func MustNewRedisAuther(host string, db *gorm.DB, rdb redis.UniversalClient, jwt *jwt.GinJWTMiddleware, opts ...OptionFunc) *Auther {
	r := &Auther{}
	r.Cache = rediscache.NewRedisStoreWithInterface(rdb)
	r.Jwt = jwt
	r.Jwt.UseAbort = true

	for _, opt := range opts {
		opt(r)
	}

	config := &fosite.Config{
		AccessTokenLifespan:  jwt.TimeoutFunc(nil),
		RefreshTokenLifespan: jwt.MaxRefreshFunc(nil),
		//授权码的有效时间
		AuthorizeCodeLifespan: 15 * time.Minute,
		IDTokenLifespan:       1 * time.Hour,

		ScopeStrategy:     fosite.WildcardScopeStrategy,
		IDTokenIssuer:     "goken",
		AccessTokenIssuer: jwt.Realm,
		//JWTScopeClaimKey: "scp",

		EnforcePKCE:                    true,
		EnforcePKCEForPublicClients:    true,
		EnablePKCEPlainChallengeMethod: false,

		//是否允许jwt里不包含jti声明
		GrantTypeJWTBearerIDOptional: false,
		//是否允许jwt里不包含iat声明
		GrantTypeJWTBearerIssuedDateOptional: false,
	}

	storage := MustNewGormClientStore(db)
	//f := fosite.NewOAuth2Provider()
	oauth2 := compose.Compose(
		config,
		storage,
		nil, // 使用默认的 HMAC 签名方法,生产环境建议换成 RSA 签名
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2PKCEFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
		compose.OAuth2RefreshTokenGrantFactory,
	)

	conf := gserver.NewConfig()
	r.OServer = gserver.MustNewServer(conf, oauth2, rdb, r.logger)
	r.Server = httpserver.MustNewServer(context.Background(), host)

	return r
}

func WithAuthenticator(auth func(*gin.Context) bool) OptionFunc {
	return func(a *Auther) {
		a.Authenticator = auth
	}
}

func WithLogger(l *otelzap.Logger) OptionFunc {
	return func(a *Auther) {
		a.logger = l
		a.Logger = l.Sugar()
	}
}

func (r *Auther) Serve() error {
	return r.Server.Serve()
}
