package jwt

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// GinJWTMiddleware 提供了一个 Json-Web-Token 认证实现,失败时返回 401 HTTP 响应,
// 成功时调用该包装的中间件后可以通过c.Get("userID").(string)获取用户ID,
// 用户可以通过向LoginHandler发送json请求来获取token,然后需要在http-header的Authentication中传递该token,
// 例如:Authorization:Bearer XXX_TOKEN_XXX
type GinJWTMiddleware struct {
	// 议题,也可以存储其他信息
	Realm string

	//受众,用于适配jwt的aud
	Audience []string

	// 签名算法-可能的值有HS256,HS384,HS512,RS256,RS384或RS512
	// 默认值HS256
	SigningAlgorithm string

	// 用于签名的密钥
	Key []byte

	// 用于动态获取签名密钥的回调函数,设置KeyFunc将绕过所有其他密钥设置,可适用于使用公私钥对的场景
	KeyFunc func(token *jwt.Token) (interface{}, error)

	// JWT token的有效时长
	Timeout time.Duration

	// 返回timeout时间,即源码内不直接使用Timeout字段,可以自定义逻辑处理Timeout
	TimeoutFunc func(data interface{}) time.Duration

	// 返回MaxRefresh时间,即源码内不直接使用MaxRefresh字段,可以自定义逻辑处理MaxRefresh
	MaxRefreshFunc func(data interface{}) time.Duration

	// 此字段允许客户端在MaxRefresh时间过去之前刷新其 token,
	// 默认为一天,传入数值为0则关闭Refresh模式
	MaxRefresh time.Duration

	//TokenInside是一个字符串,用于指定token在请求中的位置,默认值为"header",允许有多个值,用,分隔
	TokenInside string

	// TokenHeadName是header中标识Token字段的字符串,默认值为"Autorization",
	TokenHeadName string

	// TimeFunc提供当前时间,主要用于不同时区之间的连接
	//可以覆盖它以使用其他时间值,这对于测试或如果服务器使用与token不同的时区非常有用,
	TimeFunc func() time.Time

	// 是否使用gin.Context的abort(),
	UseAbort bool

	// 允许使用refresh_token,开启则强制使用security cookie(强制使用https,并且开启http only)
	EnableRefresh bool

	// 允许在验证失败时redirect,但是只允许跳转到get
	EnableRedirect bool

	// 允许使用 http.SameSite cookie参数(用于控制jwt-cookie的跨域传输),可选参数为:
	// SameSiteDefaultMode:默认行为,取决于浏览器,
	// SameSiteStrictMode:仅在同站请求中发送cookie,
	// SameSiteLaxMode:允许在跨域请求中发送cookie,
	CookieSameSite http.SameSite

	// 允许修改jwt的解析器方法
	ParseOptions []jwt.ParserOption
}

// 用于生成access_token
func (mw *GinJWTMiddleware) NewToken(keyvalue ...string) (string, time.Time, error) {
	now := mw.TimeFunc()
	claims := buildClaims(keyvalue...)
	expire := now.Add(mw.TimeoutFunc(claims))

	claims[ExpKey] = expire.Unix()
	claims[OrigIatKey] = now.Unix()
	claims[IssKey] = mw.Realm
	claims[NbfKey] = now.Add(-5 * time.Second).Unix()
	claims[IatKey] = claims[OrigIatKey]
	claims[AudKey] = mw.Audience
	claims[TokenTypeKey] = AccessTokenKey

	token := jwt.New(jwt.GetSigningMethod(mw.SigningAlgorithm))
	token.Claims = claims

	accessToken, err := token.SignedString(mw.Key)
	if err != nil {
		return "", time.Time{}, ErrFailedTokenCreation
	}
	return accessToken, expire, nil
}

// 生成refresh_token并写入HttpOnly Cookie
func (mw *GinJWTMiddleware) NewRefreshToken(keyvalue ...string) (string, error) {
	if !mw.EnableRefresh {
		return "", nil
	}

	now := mw.TimeFunc()
	claims := buildClaims(keyvalue...)

	claims[IssKey] = mw.Realm
	claims[AudKey] = mw.Audience
	claims[OrigIatKey] = now.Unix()
	claims[IatKey] = now.Unix()
	claims[NbfKey] = now.Add(-5 * time.Second).Unix()
	claims[ExpKey] = now.Add(mw.MaxRefreshFunc(claims)).Unix()
	claims[TokenTypeKey] = RefreshTokenKey

	token := jwt.New(jwt.GetSigningMethod(mw.SigningAlgorithm))
	token.Claims = claims

	refreshToken, err := token.SignedString(mw.Key)
	if err != nil {
		return "", ErrFailedTokenCreation
	}

	return refreshToken, nil
}

// 将Token加入到HttpOnly Cookie中
// TODO:开启http only
func (mw *GinJWTMiddleware) SetCookie(c *gin.Context, name, token string) {
	maxAge := int(mw.MaxRefreshFunc(nil).Seconds())
	c.SetCookie(name, token, maxAge, "/", "", false, true)

	if mw.CookieSameSite != 0 {
		c.SetSameSite(mw.CookieSameSite)
	}
}

// 辅助函数:构造基本Claims
func buildClaims(keyvalue ...string) jwt.MapClaims {
	claims := jwt.MapClaims{}
	for i := 0; i+1 < len(keyvalue); i += 2 {
		claims[keyvalue[i]] = keyvalue[i+1]
	}
	return claims
}
