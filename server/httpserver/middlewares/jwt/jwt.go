package jwt

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/hkensame/goken/pkg/common/httputil"

	"github.com/gin-gonic/gin"
)

const (
	RedirectKey     = "rediect-host"
	RefreshTokenKey = "refresh-token"
	AccessTokenKey  = "access_token"
	JwtClaimsKey    = "jwt_claims"
)

const (
	InForm   = "form"
	InJson   = "json"
	InHeader = "header"
	InQuery  = "query"
)

// 标准声明 (Standard Claims)
const (
	// 受众,标识Token的目标接收方(如:["web","mobile"])
	AudKey = "aud"
	// 过期时间(Expiration Time),Token失效的时间戳
	ExpKey = "exp"
	// Token的唯一标识(用于防止重放攻击)
	//JtiKey = "jti"
	// 签发时间(Issued At),Token签发的时间戳
	IatKey = "iat"
	// 签发者(Issuer),标识Token的签发主体(如:"auth.example.com")
	IssKey = "iss"
	// 生效时间(Not Before),在此时间之前Token无效
	NbfKey = "nbf"
	// 主题(Subject),标识Token的主体(如用户ID)
	SubKey = "sub"
	// 原始签发时间(用于刷新Token时校验)
	OrigIatKey = "orig_iat"
	// Token类型(如:"access"、"refresh")
	TokenTypeKey = "typ"
)

// gin-jwt中间件的对外使用的api,该中间件会对每一个被拦截的请求都检测jwt是否是valid
// TODO:后续估计要结合oauth来实现/login和/check
func (mw *GinJWTMiddleware) JwtAuthHandler(c *gin.Context) {
	if err := mw.authAccessToken(c); err != nil {
		if hostAny, ok := c.Get(RedirectKey); ok {
			host, _ := hostAny.(string)
			c.Redirect(http.StatusFound, host)
			return
		}
		httputil.WriteError(c, http.StatusUnauthorized, err, mw.UseAbort)
		return
	}
	c.Next()
}

// refresh会刷新超时但refresh token仍有用的请求,但注意,如果access token本身有误而不是过期,则直接返回错误
func (mw *GinJWTMiddleware) RefreshHandler(c *gin.Context) {
	if err := mw.authAccessToken(c); err == nil {
		c.Next()
		return
	} else if err == ErrExpiredToken {
		_, err := mw.authRefreshToken(c)
		if err != nil {
			if err == ErrExpiredRefreshToken {
				if hostAny, ok := c.Get(RedirectKey); ok {
					host, _ := hostAny.(string)
					c.Redirect(http.StatusFound, host)
					return
				}
			}
			httputil.WriteError(c, http.StatusUnauthorized, err, mw.UseAbort)
			return
		}
		c.Next()
	} else {
		httputil.WriteError(c, http.StatusUnauthorized, err, mw.UseAbort)
		return
	}
}

func (mw *GinJWTMiddleware) authAccessToken(c *gin.Context) error {
	tokenStr, err := mw.getToken(c)
	if err != nil {
		return err
	}

	tk, err := mw.parseTokenstr(tokenStr)
	if err != nil {
		return err
	}

	claims, ok := tk.Claims.(jwt.MapClaims)
	if !ok {
		return ErrInvalidToken
	}

	switch v := claims[ExpKey].(type) {
	case nil:
		return ErrInvalidToken
	case float64:
		if v < float64(mw.TimeFunc().Unix()) {
			return ErrExpiredToken
		}
	case int64:
		if v < mw.TimeFunc().Unix() {
			return ErrExpiredToken
		}
	default:
		return ErrInvalidToken
	}

	c.Set(JwtClaimsKey, claims)
	return nil
}

// 返回新的access_token或者错误
func (mw *GinJWTMiddleware) authRefreshToken(c *gin.Context) (string, error) {
	tokenStr, err := mw.getRefreshToken(c)
	if err != nil {
		return "", err
	}

	tk, err := mw.parseTokenstr(tokenStr)
	if err != nil {
		return "", err
	}

	claims, ok := tk.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidToken
	}

	switch v := claims[ExpKey].(type) {
	case nil:
		return "", ErrInvalidToken
	case float64:
		if v < float64(mw.TimeFunc().Unix()) {
			return "", ErrExpiredRefreshToken
		}
	case int64:
		if v < mw.TimeFunc().Unix() {
			return "", ErrExpiredRefreshToken
		}
	default:
		return "", ErrInvalidToken
	}

	kv := make([]string, 0, len(claims)*2)
	//只有newToken函数会写入claims,而伪造或错误的claims不会执行到这一行,这里就能保证v一定是string
	for k, v := range claims {
		sv, ok := v.(string)
		if !ok {
			continue
		}
		kv = append(kv, k, sv)
	}
	acctk, _, err := mw.NewToken(kv...)
	if err != nil {
		return "", err
	}

	c.Header(AccessTokenKey, acctk)
	c.Set(JwtClaimsKey, claims)
	return acctk, nil

}
