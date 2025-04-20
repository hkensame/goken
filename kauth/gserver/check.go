package gserver

import (
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
)

/*
	check工作应该在三部分进行:
	1.在创建时应当通过manager检查待加入clientInfo的格式是否正确
	2.在/authorize请求中检查请求体是否和记录的clientInfo一样
	3.在/token请求中检查请求体是否有权利获得token
*/

func (s *Server) CheckAllowedClientRequst(req *AuthorizedInfo) (err error) {
	infoAny, _ := req.Ctx.Get(ClientInfoStr)
	info, ok := infoAny.(ClientInfo)
	if !ok {
		return errors.ErrInvalidRequest
	}

	if !s.checkResponseType(req.ResponseType.String(), info) {
		return errors.ErrUnsupportedResponseType
	}

	if !s.checkRedirectURL(req.RedirectURI) {
		return errors.ErrInvalidRedirectURI
	}

	if !s.checkScope(req.Scope, info) {
		return errors.ErrInvalidScope
	}

	if err := s.checkCodeChallenge(req); err != nil {
		return err
	}
	return nil
}

func (s *Server) CheckAllowedTokenRequest(tk *TokenInfo) error {
	if tk.GrantType == oauth2.Code.String() {
		if tk.Code == "" {
			return errors.ErrInvalidAuthorizeCode
		}
		if !s.checkRedirectURL(tk.RedirectURI) {
			return errors.ErrInvalidRedirectURI
		}

	} else if tk.GrantType == oauth2.ClientCredentials.String() {
		if tk.ClientSecret == "" {
			return errors.ErrInvalidRequest
		}
	} else {
		if tk.RefreshToken == "" {
			return errors.ErrInvalidRefreshToken
		}
	}

	if s.Config.ForcePKCE && tk.CodeVerifier == "" {
		return errors.ErrMissingCodeVerifier
	}

	//这里应该根据不同的模式到不同的位置查询?
	if tk.Scope != "" {
	}

	return nil
}

func (s *Server) checkHttpMethod(method string) error {
	if !(method == "POST" || (s.Config.AllowGetAccessRequest && method == "GET")) {
		return errors.ErrInvalidRequest
	}
	return nil
}

// TODO: 后续需要逻辑定义,解析req中scopes里的格式和内容
// 暂时只处理这样的格式: email,mobile,nickname或email mobile nickname
func (s *Server) checkScope(scopes string, info ClientInfo) bool {
	scopes = strings.ReplaceAll(strings.TrimSpace(strings.ToLower(scopes)), ",", " ")
	fields := strings.Fields(scopes)
	elems := make(map[string]struct{})
	for _, v := range info.GetScopes() {
		elems[v] = struct{}{}
	}
	for _, v := range fields {
		if _, ok := elems[v]; !ok {
			return false
		}
	}
	return true
}

func (s *Server) checkResponseType(rt string, info ClientInfo) bool {
	for _, v := range info.GetResponseTypes() {
		if v == rt {
			return true
		}
	}
	return false
}

// 检查请求中是否是允许的CodeChallengeMethod
func (s *Server) checkCodeChallengeMethod(ccm oauth2.CodeChallengeMethod) bool {
	for _, c := range s.Config.AllowedCodeChallengeMethods {
		if c == ccm {
			return true
		}
	}
	return false
}

func (s *Server) checkCodeChallenge(req *AuthorizedInfo) error {
	if s.Config.ForcePKCE && req.CodeChallenge == "" {
		return errors.ErrCodeChallengeRquired
	}

	// 非强制PKCE时,如果未提供code_challenge则直接跳过验证
	if req.CodeChallenge == "" {
		return nil
	}

	if len(req.CodeChallenge) < 43 || len(req.CodeChallenge) > 128 {
		return errors.ErrInvalidCodeChallengeLen
	}

	if req.CodeChallengeMethod == "" || (req.CodeChallengeMethod != oauth2.CodeChallengePlain &&
		req.CodeChallengeMethod != oauth2.CodeChallengeS256) {
		return errors.ErrUnsupportedCodeChallengeMethod
	}

	if req.CodeChallengeMethod == oauth2.CodeChallengeS256 {
		if _, err := base64.RawURLEncoding.DecodeString(req.CodeChallenge); err != nil {
			return errors.ErrInvalidCodeChallenge
		}
	}

	if err := s.checkAccessType(req.AccessType); err != nil {
		return err
	}
	if err := s.checkState(req.AccessType); err != nil {
		return err
	}
	if err := s.checkPrompt(req.AccessType); err != nil {
		return err
	}

	return nil
}

func (s *Server) checkState(state string) error {
	if len(state) < 8 || len(state) > 128 {
		return errors.ErrInvalidRequest
	}
	return nil
}

func (s *Server) checkPrompt(prompt string) error {
	if prompt == "" || validPrompts[prompt] {
		return nil
	}
	return errors.ErrInvalidRequest
}

func (s *Server) checkAccessType(accessType string) error {
	if accessType == "" || validAccessType[accessType] {
		return nil
	}
	return errors.ErrInvalidRequest
}

func (s *Server) checkRedirectURL(us string) bool {
	_, err := url.Parse(us)
	if err != nil {
		return false
	}
	return true
}
