package gserver

import (
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/hkensame/goken/pkg/common/httputil"
)

// 从内部数据库中获取clientID的信息
func (s *Server) GetClientInfo(ctx *gin.Context, clientID string) (cli oauth2.ClientInfo, err error) {
	res, err := s.Manager.GetClient(ctx.Request.Context(), clientID)
	if err != nil {
		return nil, err
	}
	if _, ok := res.(ClientInfo); !ok {
		return nil, errors.ErrInvalidClient
	}
	ctx.Set(ClientInfoStr, res)
	return res, err
}

// 从/authorize请求中获取AuthorizedInfo信息
func (s *Server) GetAuthorizedInfo(c *gin.Context) (*AuthorizedInfo, error) {
	res := &AuthorizedInfo{}
	if !httputil.MustIsMethod(c, "GET") {
		return nil, errors.ErrInvalidRequest
	}
	if err := c.ShouldBindQuery(res); err != nil {
		return nil, errors.ErrInvalidRequest
	}

	res.Ctx = c
	return res, nil
}

// 从/token请求中获取TokenInfo信息
func (s *Server) GetTokenInfo(c *gin.Context) (*TokenInfo, error) {
	res := &TokenInfo{}
	if err := s.checkGetTokenMethod(c.Request.Method); err != nil {
		return nil, err
	}
	if err := c.ShouldBind(res); err != nil {
		return nil, errors.ErrInvalidRequest
	}

	res.Ctx = c
	return res, nil
}

func (s *Server) checkGetTokenMethod(method string) error {
	if !(method == "POST" || (s.Config.AllowGetAccessRequest && method == "GET")) {
		return errors.ErrInvalidRequest
	}
	return nil
}
