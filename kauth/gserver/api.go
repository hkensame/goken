package gserver

import (
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/hkensame/goken/pkg/common/httputil"
)

func (s *Server) addClientUrlInfo(req *AuthorizedInfo) string {
	//check时已经检验过一次
	u, _ := url.Parse(req.RedirectURI)
	u.Query().Add("state", req.State)
	//u.Query().Add()
	return u.String()
}

func (s *Server) Authorize() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := s.GetAuthorizedInfo(ctx)

		if err != nil {
			//此时直接返回,不能确定url是有效的
			httputil.WriteError(ctx, errors.StatusCodes[err], err, true)
			return
		}
		if err := s.CheckAllowedClientRequst(res); err != nil {
			httputil.WriteError(ctx, errors.StatusCodes[err], err, true)
			return
		}

		res.RedirectURI = s.addClientUrlInfo(res)

		switch res.ResponseType.String() {
		case oauth2.Code.String():
			tk, err := s.generateCode(res)
			if err != nil {
				ctx.Redirect(errors.StatusCodes[err], res.RedirectURI)
				return
			}
			code := tk.GetCode()
			store := code + ":" + res.CodeChallenge + ":" + string(res.CodeChallengeMethod)
			fmt.Println(store)
		case oauth2.Token.String():
		case IDToken:
		}
	}
}
