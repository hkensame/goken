package gserver

import (
	"encoding/base64"
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

// 对收到的AuthorizedInfo的初步检查,与数据库底层记录的数据是否一致
func (s *Server) CheckAllowedAuthorizeRequst(req *AuthorizedInfo) error {
	infoAny, _ := req.Ctx.Get(ClientInfoStr)
	info, ok := infoAny.(ClientInfo)
	if !ok {
		return errors.ErrInvalidRequest
	}

	var err error
	if err = s.checkResponseType(req.ResponseType.String(), info); err != nil {
		return err
	}

	if err = s.checkRedirectURL(req.RedirectURI, info); err != nil {
		return err
	}

	if err = s.checkScope(req, info); err != nil {
		return err
	}

	if err = s.checkCodeChallenge(req, info); err != nil {
		return err
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
	req.IsPublic = info.IsPublic()
	return nil
}

// TODO: 后续需要逻辑定义,解析req中scopes里的格式和内容
// 暂时只处理这样的格式: email,mobile,nickname或email mobile nickname
func (s *Server) checkScope(req *AuthorizedInfo, info ClientInfo) error {
	req.Scope = strings.ReplaceAll(strings.TrimSpace(strings.ToLower(req.Scope)), ",", " ")
	fields := strings.Fields(req.Scope)
	elems := make(map[string]struct{})
	for _, v := range info.GetScopes() {
		elems[v] = struct{}{}
	}
	for _, v := range fields {
		if _, ok := elems[v]; !ok {
			return errors.ErrInvalidScope
		}
	}
	return nil
}

func (s *Server) checkResponseType(rt string, info ClientInfo) error {
	for _, v := range info.GetResponseTypes() {
		if v == rt {
			return nil
		}
	}
	return errors.ErrUnsupportedResponseType
}

func (s *Server) checkCodeChallenge(req *AuthorizedInfo, info ClientInfo) error {
	// TODO: 这里先不考虑 id_token 模式
	if !s.Config.UsePKCE {
		return nil
	}

	//使用了/authorize接口且UsePKCE则必定意味着要带code-challenge信息
	if (req.CodeChallenge == "" || req.CodeChallengeMethod == "") && (info.IsPublic() || s.Config.UsePKCE) {
		return errors.ErrCodeChallengeRquired
	}

	if req.CodeChallengeMethod != oauth2.CodeChallengeS256 {
		return errors.ErrUnsupportedCodeChallengeMethod
	}

	if len(req.CodeChallenge) < 43 || len(req.CodeChallenge) > 128 {
		return errors.ErrInvalidCodeChallengeLen
	}

	//默认客户端生成pkce时会对code-challenge进行一次base64序列化,这里测试看是否是base64格式
	if _, err := base64.RawURLEncoding.DecodeString(req.CodeChallenge); err != nil {
		return errors.ErrInvalidCodeChallenge
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
	return errors.ErrAccessDenied
}

func (s *Server) checkRedirectURL(us string, info ClientInfo) error {
	for _, v := range info.GetRedirectURIs() {
		if us == v {
			return nil
		}
	}
	return errors.ErrInvalidRedirectURI
}

/*
	Token Check函数
*/

// func (s *Server) checkClientSercet(req *AuthorizedInfo, info ClientInfo) error {
// 	//使用code不需要验证sercet
// 	if req.ResponseType == oauth2.Code {
// 		return nil
// 	}
// 	if req.ClientSecret == info.GetSecret() {
// 		return nil
// 	}
// 	return errors.ErrInvalidRequest
// }
