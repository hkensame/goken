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
		Timeout:          time.Hour * 2,
		MaxRefresh:       time.Hour * 48,
		Realm:            "gin-jwt",
		UseAbort:         true,
		TimeFunc:         time.Now,
		Key:              []byte(key),
		CookieSameSite:   http.SameSiteDefaultMode,
		EnableRefresh:    true,
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
	ErrMissRefreshToken = errors.New("miss refresh token")
	ErrMissToken        = errors.New("miss token")
	// 表示缺少密钥
	ErrMissingSecretKey = errors.New("secret key is required")

	// 表示JWT令牌创建失败,原因未知
	ErrFailedTokenCreation = errors.New("failed to create JWT Token")

	// 表示JWT令牌已过期,无法刷新
	ErrExpiredToken = errors.New("token is expired")

	// 表示刷新令牌已过期,无法刷新
	ErrExpiredRefreshToken = errors.New("refresh token is expired")

	// 表示token无效
	ErrInvalidToken = errors.New("token is invalid")

	// 表示签名算法无效,必须是 HS256、HS384、HS512、RS256、RS384 或 RS512
	ErrInvalidSigningAlgorithm = errors.New("invalid signing algorithm")
)

var (
	ErrEmptyQueryToken  = errors.New("query token is empty")
	ErrEmptyFormToken   = errors.New("form token is empty")
	ErrEmptyHeaderToken = errors.New("header token is empty")
	ErrEmptyJsonToken   = errors.New("json token is empty")
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

func WithKeyFunc(keyFunc func(t *jwt.Token) (interface{}, error)) JwtOption {
	return func(mw *GinJWTMiddleware) {
		mw.KeyFunc = keyFunc
	}
}
