package authmodel

import (
	"github.com/gin-gonic/gin"
)

func (r *Auther) AddClientInfoHandler(c *gin.Context) {

}

// 如果你还打算支持 OIDC，可以新增：
// /userinfo：返回 sub, email, name 等
// /jwks：返回公开 JWK（方便验证 token）
// /well-known/openid-configuration
// 如果你想我可以帮你加上这些接口。
