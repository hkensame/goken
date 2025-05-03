package jwt

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func (mw *GinJWTMiddleware) getRefreshToken(c *gin.Context) (string, error) {
	if rfs, err := c.Cookie(RefreshTokenKey); err != nil {
		return "", ErrMissRefreshToken
	} else {
		return rfs, nil
	}
}

// 从gin.Context中获取jwt.Token
func (mw *GinJWTMiddleware) getToken(c *gin.Context) (string, error) {
	var token string
	var err error
	inside := strings.Split(mw.TokenInside, ",")
	for _, v := range inside {
		switch v {
		case InHeader:
			token, err = mw.jwtFromHeader(c, mw.TokenHeadName)
		case InQuery:
			token, err = mw.jwtFromQuery(c, mw.TokenHeadName)
		case InForm:
			token, err = mw.jwtFromForm(c, mw.TokenHeadName)
		case InJson:
			token, err = mw.jwtFromJson(c, mw.TokenHeadName)
		}
		if token != "" || err != nil {
			break
		}
	}

	if err != nil {
		return "", err
	}
	return token, nil
}

// 将TokenStr反序列为jwt.Token
func (mw *GinJWTMiddleware) parseTokenstr(token string) (*jwt.Token, error) {
	var tk *jwt.Token
	var err error

	if mw.KeyFunc != nil {
		tk, err = jwt.Parse(token, mw.KeyFunc, mw.ParseOptions...)
	} else {
		tk, err = jwt.Parse(token,
			func(t *jwt.Token) (interface{}, error) {
				if jwt.GetSigningMethod(mw.SigningAlgorithm) != t.Method {
					return nil, ErrInvalidSigningAlgorithm
				}
				return mw.Key, nil
			},
			mw.ParseOptions...,
		)
	}

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, ErrExpiredToken
			}
		}
		return nil, ErrInvalidToken
	}

	return tk, nil
}
func (mw *GinJWTMiddleware) jwtFromHeader(c *gin.Context, key string) (string, error) {
	token := c.Request.Header.Get(key)
	if token == "" {
		return "", ErrEmptyHeaderToken
	}
	return token, nil
}

func (mw *GinJWTMiddleware) jwtFromJson(c *gin.Context, key string) (string, error) {
	tk := struct {
		Token string `json:"key"`
	}{}
	if err := c.ShouldBindBodyWithJSON(&tk); err != nil {
		return "", ErrEmptyJsonToken
	}
	return tk.Token, nil
}

func (mw *GinJWTMiddleware) jwtFromForm(c *gin.Context, key string) (string, error) {
	token := c.PostForm(key)
	if token == "" {
		return "", ErrEmptyFormToken
	}
	return token, nil
}

func (mw *GinJWTMiddleware) jwtFromQuery(c *gin.Context, key string) (string, error) {
	token := c.Query(key)
	if token == "" {
		return "", ErrEmptyQueryToken
	}
	return token, nil
}
