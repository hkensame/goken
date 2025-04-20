package authmodel

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/hkensame/goken/pkg/common/httputil"
	"github.com/hkensame/goken/server/httpserver/middlewares/jwt"
)

func (r *Auther) AuthorizeHandler(c *gin.Context) {
	err := r.OServer.HandleAuthorizeRequest(c.Writer, c.Request)
	if err != nil {
		httputil.WriteResponse(c, 400, "", err, true)
	}
}

func (r *Auther) NewTokenHandler(c *gin.Context) {
	err := r.OServer.HandleTokenRequest(c.Writer, c.Request)
	if err != nil {
		httputil.WriteResponse(c, 400, "", err, true)
	}
}

func (r *Auther) AddClientInfoHandler(c *gin.Context) {

}

// 如果你还打算支持 OIDC，可以新增：
// /userinfo：返回 sub, email, name 等
// /jwks：返回公开 JWK（方便验证 token）
// /well-known/openid-configuration
// 如果你想我可以帮你加上这些接口。

// 注册中间件,暂时只能使用开发者提供的认证函数
func (r *Auther) LoginHandler(ctx *gin.Context) {
	if ok := r.Authenticator(ctx); ok {
		tokenString, expire, err := r.Jwt.NewToken()
		if err != nil {
			httputil.WriteResponse(ctx, http.StatusUnauthorized, err.Error(), nil, r.Jwt.UseAbort)
		}
		r.Jwt.SetCookie(ctx, tokenString)
		httputil.WriteResponse(ctx, http.StatusOK, "", gin.H{"token": tokenString, "expire": expire}, false)
	} else {
		httputil.WriteResponse(ctx, http.StatusUnauthorized, ErrFailedAuthentication.Error(), nil, r.Jwt.UseAbort)
	}
}

// LogoutHandler作为Middleware可供客户端使用,用于移除jwt的cookie
func (r *Auther) LogoutHandler(c *gin.Context) {
	if r.Jwt.SendCookie {
		if r.Jwt.CookieSameSite != 0 {
			c.SetSameSite(r.Jwt.CookieSameSite)
		}

		//通过设置空值移除cookie
		c.SetCookie(
			r.Jwt.CookieName,
			"",
			-1,
			"/",
			"",
			r.Jwt.SecureCookie,
			r.Jwt.SecureCookie,
		)
	}
}

// 外层授权,这里只针对客户端代理,注意这里不是用户授权
func (r *Auther) authenticationHandler(c *gin.Context) {
	rediretHost := "/sso/login?redirect=" + url.QueryEscape(c.Request.RequestURI)
	claims, err := r.Jwt.GetClaimsFromContext(c)
	if err != nil {
		c.Redirect(http.StatusFound, rediretHost)
	}

	//如果存在clientId则认可其身份
	clientId, exists := claims["aud"].(string)
	if !exists || clientId == "" {
		c.Redirect(http.StatusFound, rediretHost)
	}

	//以便于后续如果token过期时能找到redirectHost
	c.Set(jwt.RedirectHost, rediretHost)
}

// nonce 检查
func (r *Auther) RouterRegister() {
	r.Server.Engine.GET("/authorize", r.AuthorizeHandler)
	r.Server.Engine.POST("/token", r.NewTokenHandler)
	r.Server.Engine.POST("/clientinfos", r.AddClientInfoHandler)
	r.Server.Engine.GET("/authenticate", r.authenticationHandler, r.Jwt.RefreshHandler, r.Jwt.JwtAuthHandler)
}
