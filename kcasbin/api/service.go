package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hkensame/goken/kcasbin/proto"
	kcasbin "github.com/hkensame/goken/kcasbin/server"
)

func main() {
	r := gin.Default()

	kasbin := kcasbin.MustNewGormKasbin(nil /* 加载db、logger、watcher等 */)

	r.POST("/policy/add", func(c *gin.Context) {
		var req proto.MatchPolicies
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := kasbin.AddPolicies(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	r.POST("/policy/remove", func(c *gin.Context) {
		var req proto.MatchPolicies
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err := kasbin.RemovePolicies(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	r.POST("/auth", func(c *gin.Context) {
		var req proto.AuthorizeReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := kasbin.Authorize(c.Request.Context(), &req)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"ok": res.GetOk(), "detail": res.GetDetail()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": res.GetOk(), "detail": res.GetDetail()})
	})

	r.GET("/roles/:user", func(c *gin.Context) {
		user := c.Param("user")
		res, err := kasbin.GetUserRoles(c.Request.Context(), &proto.GetUserRolesReq{User: user})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/users/:role", func(c *gin.Context) {
		role := c.Param("role")
		res, err := kasbin.GetRoleUsers(c.Request.Context(), &proto.GetRoleUsersReq{Role: role})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.Run(":8080")
}
