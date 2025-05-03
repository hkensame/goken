package gserver

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-redis/cache/v9"
	"github.com/hkensame/goken/pkg/common/httputil"
	jsoniter "github.com/json-iterator/go"
)

func (s *Server) addUrlInfo(u *url.URL, key string, val string) *url.URL {
	q := u.Query()
	q.Add(key, val)
	u.RawQuery = q.Encode()
	return u
}

const (
	CodeKey                = "code"
	CodeChallengeKey       = "cc"
	CodeChallengeMethodKey = "ccm"
	ClientIDKey            = "cid"
	RedirectURLKey         = "rdru"
	CreatedAtKey           = "ct"
)

type authCodeData struct {
	Code                string `json:"code"`
	CodeChallenge       string `json:"cc,omitempty"`
	CodeChallengeMethod string `json:"ccm,omitempty"`
	ClientID            string `json:"cid"`
	RedirectURI         string `json:"rdru"`
	CreatedAt           int64  `json:"ct"`
	IsPublic            bool   `json:"pub"`
	Scope               string `json:"scp"`
}

/*
	注意这里没有考虑code challenge时code/client_id无效,许把code challenge作为code
*/

func (s *Server) Authorize() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := s.GetAuthorizedInfo(ctx)
		if err != nil {
			// 无法确定RedirectURI是否有效,直接返回错误
			httputil.WriteError(ctx, errors.StatusCodes[err], err, true)
			return
		}

		if _, err := s.GetClientInfo(ctx, res.ClientID); err != nil {
			httputil.WriteError(ctx, errors.StatusCodes[err], err, true)
		}
		if err := s.CheckAllowedAuthorizeRequst(res); err != nil {
			httputil.WriteError(ctx, errors.StatusCodes[err], err, true)
			return
		}
		// 检查时已经通过,不用担心出错
		u, _ := url.Parse(res.RedirectURI)
		u = s.addUrlInfo(u, "state", res.State)

		switch res.ResponseType.String() {
		case oauth2.Code.String():
			tk, err := s.generateCode(res)
			if err != nil {
				u = s.addUrlInfo(u, "error", errors.ErrServerError.Error())
				ctx.Redirect(http.StatusFound, u.String())
				return
			}
			code := tk.GetCode()
			key := fmt.Sprintf("kauth:code:%s", code)

			codeData := &authCodeData{
				Code:        code,
				ClientID:    res.ClientID,
				RedirectURI: res.RedirectURI,
				CreatedAt:   time.Now().Unix(),
				IsPublic:    res.IsPublic,
			}

			if s.Config.UsePKCE {
				codeData.CodeChallenge = res.CodeChallenge
				codeData.CodeChallengeMethod = string(res.CodeChallengeMethod)
			}

			val, err := jsoniter.Marshal(codeData)
			if err != nil {
				s.Logger.Errorf("[kauth] json encode 失败: %v", err)
				u = s.addUrlInfo(u, "error", errors.ErrServerError.Error())
				ctx.Redirect(http.StatusFound, u.String())
				return
			}

			err = s.Cache.Set(&cache.Item{
				Ctx:   ctx.Request.Context(),
				Key:   key,
				Value: string(val),
				TTL:   s.Config.CodeTTL,
			})
			if err != nil {
				s.Logger.Errorf("[kauth] cache set失败, err = %v", err)
				u = s.addUrlInfo(u, "error", errors.ErrServerError.Error())
				ctx.Redirect(http.StatusFound, u.String())
				return
			}

			// 成功返回 code
			u = s.addUrlInfo(u, "code", code)
			ctx.Redirect(http.StatusFound, u.String())

		case IDTokenSecure:
			// 如果你要实现 hybrid/implicit 流程，这里处理 id_token
			httputil.WriteError(ctx, http.StatusBadRequest, errors.ErrUnsupportedResponseType, true)
		default:
			u = s.addUrlInfo(u, "error", "unsupported_response_type")
			ctx.Redirect(http.StatusFound, u.String())
		}
	}
}

func (s *Server) Token() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := s.GetTokenInfo(ctx)
		if err != nil {
			httputil.WriteError(ctx, errors.StatusCodes[err], err, true)
			return
		}

		switch res.GrantType {
		case oauth2.AuthorizationCode.String():
			if res.Code == "" {
				err = errors.ErrInvalidAuthorizeCode
				httputil.WriteError(ctx, errors.StatusCodes[err], err, true)
				return
			}
			if res.ClientSecret == "" {
				err = errors.ErrInvalidRequest
				httputil.WriteError(ctx, errors.StatusCodes[err], err, true)
				return
			}
			val := []byte{}
			key := fmt.Sprintf("kauth:code:%s", res.Code)

			if err := s.Cache.Get(ctx.Request.Context(), key, &val); err != nil {
				httputil.WriteError(ctx, http.StatusBadRequest, errors.ErrInvalidAuthorizeCode, true)
				return
			}

			codeData := &authCodeData{}
			jsoniter.Unmarshal(val, codeData)

			if err := s.CheckValidCodeTokenInfo(codeData, res); err != nil {
				httputil.WriteError(ctx, http.StatusBadRequest, err, true)
				return
			}

			gt, err := s.generateToken(res)
			if err != nil {
				fmt.Println("wwwwwwwqwqiwgnqoiwngiqonegoiqngiow")
				httputil.WriteError(ctx, http.StatusInternalServerError, err, true)
				return
			}

			httputil.WriteResponse(ctx, http.StatusOK, "", gin.H{
				"access_token":  gt.GetAccess(),
				"token_type":    s.Config.TokenType,
				"expires_in":    gt.GetAccessExpiresIn(),
				"refresh_token": gt.GetRefresh(),
			}, true)
			return
			// case oauth2.ClientCredentials.String():
			// 	if tk.ClientSecret == "" {
			// 		return errors.ErrInvalidRequest
			// 	}
			// case oauth2.Refreshing.String():
			// 	if tk.RefreshToken == "" {
			// 		return errors.ErrInvalidRefreshToken
			// 	}
		}

		httputil.WriteError(ctx, http.StatusBadRequest, errors.ErrUnsupportedGrantType, true)
	}
}

func (s *Server) Introspect(f gin.HandlerFunc) gin.HandlerFunc {
	return f
}
