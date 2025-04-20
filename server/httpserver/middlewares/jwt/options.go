package jwt

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JwtOption func(*GinJWTMiddleware)

func MustNewGinJWTMiddleware(key string, opts ...JwtOption) *GinJWTMiddleware {
	mw := &GinJWTMiddleware{
		TokenInside:      "header",
		TokenHeadName:    "Authorization",
		SigningAlgorithm: "HS256",
		Timeout:          time.Hour * 24,
		MaxRefresh:       time.Hour * 24,
		Realm:            "gin-jwt",
		CookieMaxAge:     time.Hour * 24,
		CookieName:       "jwt-cookie",
		ExpField:         "exp",
		SendCookie:       false,
		SecureCookie:     true,
		UseAbort:         true,
		TimeFunc:         time.Now,
		Key:              []byte(key),
		CookieSameSite:   http.SameSiteDefaultMode,
	}

	// 应用所有选项函数
	for _, JwtOption := range opts {
		JwtOption(mw)
	}

	if mw.TimeoutFunc == nil {
		mw.TimeoutFunc = func(data interface{}) time.Duration {
			return mw.Timeout
		}
	}

	if mw.MaxRefreshFunc == nil {
		mw.MaxRefreshFunc = func(data interface{}) time.Duration {
			return mw.MaxRefresh
		}
	}

	//如果有能获得Key的Func就无需再判断有没有Key了
	if mw.KeyFunc != nil {
		return mw
	}

	if mw.Key == nil {
		panic(ErrMissingSecretKey)
	}

	return mw
}

var (
	// ErrMissingSecretKey 表示缺少密钥
	ErrMissingSecretKey = errors.New("secret key is required")

	// ErrFailedTokenCreation 表示JWT令牌创建失败,原因未知
	ErrFailedTokenCreation = errors.New("failed to create JWT Token")

	// ErrExpiredToken 表示JWT令牌已过期,无法刷新
	ErrExpiredToken = errors.New("token is expired")

	// ErrExpiredRefreshToken 表示刷新令牌已过期,无法刷新
	ErrExpiredRefreshToken = errors.New("refresh token is expired")

	// ErrMissingExpField 表示令牌缺少 `exp`（过期时间）字段
	ErrMissingExpField = errors.New("missing exp field")

	// ErrInvalidToken 表示token无效
	ErrInvalidToken = errors.New("表示token无效 is invalid")

	// ErrInvalidSigningAlgorithm 表示签名算法无效,必须是 HS256、HS384、HS512、RS256、RS384 或 RS512
	ErrInvalidSigningAlgorithm = errors.New("invalid signing algorithm")

	// ErrTokenStringInvalid 表示tokenstring是无效的
	ErrTokenStringInvalid = errors.New("token string invalid")
)

var (

	// ErrEmptyQueryToken 在 URL 查询参数或Form Data中进行认证时,如果令牌变量为空,则会抛出该错误
	ErrEmptyQueryToken = errors.New("query token is empty")

	// ErrEmptyCookieToken 在使用 Cookie 进行认证时,如果令牌 Cookie 为空,则会抛出该错误
	ErrEmptyCookieToken = errors.New("cookie token is empty")

	// ErrEmptyParamToken 在使用 URL 路径参数进行认证时,如果参数为空,则会抛出该错误
	ErrEmptyParamToken = errors.New("parameter token is empty")

	// ErrEmptyHeadToken 在使用 HTTP 头部进行认证时,如果 `Authorization` 头为空,则会抛出该错误
	ErrEmptyHeadToken = errors.New("token header is empty")
)

func WithTokenInside(t string) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.TokenInside = t
	}
}

func WithTokenHeadName(h string) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.TokenHeadName = h
	}
}

func WithSigningAlgorithm(algorithm string) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.SigningAlgorithm = algorithm
	}
}

func WithTimeout(timeout time.Duration) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.Timeout = timeout
	}
}

func WithTimeoutFunc(timeoutFunc func(data interface{}) time.Duration) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.TimeoutFunc = timeoutFunc
	}
}

func WithTimeFunc(timeFunc func() time.Time) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.TimeFunc = timeFunc
	}
}

// func WithHTTPStatusMessageFunc(httpStatusMessageFunc func(e error, c *gin.Context) string) JwtOption {
// 	return func(mw *GinJWTMiddleware) {
// 		mw.HTTPStatusMessageFunc = httpStatusMessageFunc
// 	}
// }

func WithRealm(realm string) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.Realm = realm
	}
}

func WithCookieMaxAge(cookieMaxAge time.Duration) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.CookieMaxAge = cookieMaxAge
	}
}

func WithCookieName(cookieName string) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.CookieName = cookieName
	}
}

func WithExpField(expField string) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.ExpField = expField
	}
}

func WithKeyFunc(keyFunc func(t *jwt.Token) (interface{}, error)) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.KeyFunc = keyFunc
	}
}
