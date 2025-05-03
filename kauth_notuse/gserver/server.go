package gserver

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-redis/cache/v9"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

const (
	AuthorizedInfoStr = "authorized-info"
	ClientInfoStr     = "client-info"
)

const (
	IDTokenSecure = "code id_token"
)

const (
	PKCEKey = "pkce:"
	CCKey   = "cck:"
	CCMKey  = "ccmk:"
	URLKey  = "rdr-url:"
)

var validPrompts = map[string]bool{
	"none":           true,
	"login":          true,
	"consent":        true,
	"select_account": true,
}

var validAccessType = map[string]bool{
	"online":  true, // 默认值,返回短期有效的access_token
	"offline": true, // 可返回refresh_token
}

// AuthorizedInfo 表单绑定结构体
type AuthorizedInfo struct {
	ResponseType        oauth2.ResponseType        `form:"response_type" binding:"required,oneof=token id_token code"`
	ClientID            string                     `form:"client_id" binding:"required"`
	RedirectURI         string                     `form:"redirect_uri" binding:"required"`
	Scope               string                     `form:"scope" binding:"required"`
	State               string                     `form:"state" binding:"required"`
	CodeChallenge       string                     `form:"code_challenge"`
	CodeChallengeMethod oauth2.CodeChallengeMethod `form:"code_challenge_method"`
	//与oidc有关
	Prompt     string `form:"prompt"`
	AccessType string `form:"access_type"`
	//不允许支持implicit,不需要nonce
	//nonce  string
	IsPublic bool         `form:"-"`
	UserID   string       `form:"-"`
	Ctx      *gin.Context `form:"-"`
}

type TokenInfo struct {
	// 只允许 "authorization_code", "client_credentials", "refresh_token"
	GrantType string `form:"grant_type" binding:"required,oneof=authorization_code client_credentials refresh_token"`
	ClientID  string `form:"client_id" binding:"required"`
	// 使用client_credentials模式时必填
	ClientSecret string `form:"client_secret"`
	// 使用code模式时以下两个字段必填
	Code        string `form:"code"`
	RedirectURI string `form:"redirect_uri"`
	// 若授权请求使用了code_challenge,则必填
	CodeVerifier string `form:"code_verifier"`
	Scope        string `form:"scope"`
	// 使用refresh_token模式时必填
	RefreshToken string `form:"refresh_token"`
	// 可选自定义access_token过期时间,但只能比默认的过期时间的小,暂时不知道该字段的意义是什么
	AccessTokenExp time.Duration `form:"access_token_exp" binding:"omitempty,lt=86400"`
	UserID         string        `form:"-"`
	Ctx            *gin.Context  `form:"-"`
	//作为标识是否需要PKCE,在check函数中填充
	needPKCE bool
}

type Server struct {
	Config  *config
	Manager *manage.Manager //oauth2.Manager
	Cache   *cache.Cache
	Logger  *otelzap.SugaredLogger
	logger  *otelzap.Logger
}

type config struct {
	//tokenType标识生成的token存储的前缀,默认为Bearer
	TokenType string
	//tokenPlace标识生成的token存储的位置
	TokenPlace string
	//标识是否允许GET方法获取token
	AllowGetAccessRequest bool
	//标识允许的ResponseType
	AllowedResponseTypes []oauth2.ResponseType
	// 标识允许的GrantType
	AllowedGrantTypes []oauth2.GrantType
	// 标识允许的GCodeChallengeMethods
	AllowedCodeChallengeMethods []oauth2.CodeChallengeMethod
	//要求public客户端或单页前端使用pkce,暂时隐蔽字段不允许关闭pkce
	UsePKCE bool
	//code模式下,code的时效时间,默认为10分钟
	CodeTTL time.Duration
}

func MustNewServer(cfg *config, manager *manage.Manager, cli redis.UniversalClient, l *otelzap.Logger) *Server {
	srv := &Server{
		Config:  cfg,
		Manager: manager,
		Cache: cache.New(&cache.Options{
			Redis: cli,
		}),
		logger: l,
		Logger: l.Sugar(),
	}
	return srv
}

// 默认配置中只允许使用AuthorizationCode,ClientCredentials,Refreshing三种GrantTypes
// 且只允许ClientCredentials使用Token ResponseTypes
// 默认也只支持CodeChallengeS256一种模式,
// 不允许GET方法获取Token,不强制使用pkce
func NewConfig() *config {
	return &config{
		TokenType:            "Bearer",
		TokenPlace:           "header",
		AllowedResponseTypes: []oauth2.ResponseType{oauth2.Code, IDTokenSecure},
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.ClientCredentials,
			oauth2.Refreshing,
		},
		AllowedCodeChallengeMethods: []oauth2.CodeChallengeMethod{
			oauth2.CodeChallengeS256,
		},
		CodeTTL: 10 * time.Minute,
		UsePKCE: true,
	}
}

// 扩展oauth2原有的ClientInfo
type ClientInfo interface {
	GetID() string
	GetSecret() string
	GetDomain() string
	IsPublic() bool
	GetUserID() string
	GetGrantTypes() []string
	GetScopes() []string
	GetRedirectURIs() []string
	GetResponseTypes() []string
}

func (s *Server) SetTokenType(tokenType string) {
	s.Config.TokenType = tokenType
}

func (s *Server) SetAllowGetAccessRequest(allow bool) {
	s.Config.AllowGetAccessRequest = allow
}

func (s *Server) SetAllowedResponseType(types ...oauth2.ResponseType) {
	s.Config.AllowedResponseTypes = types
}

func (s *Server) SetAllowedGrantType(types ...oauth2.GrantType) {
	s.Config.AllowedGrantTypes = types
}
