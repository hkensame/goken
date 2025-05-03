package gserver

import (
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
)

func (s *Server) generateCode(req *AuthorizedInfo) (authToken oauth2.TokenInfo, err error) {
	tgr := &oauth2.TokenGenerateRequest{
		ClientID:            req.ClientID,
		Scope:               req.Scope,
		RedirectURI:         req.RedirectURI,
		Request:             req.Ctx.Request,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: req.CodeChallengeMethod,
		UserID:              req.UserID,
	}
	res, err := s.Manager.GenerateAuthToken(req.Ctx.Request.Context(), req.ResponseType, tgr)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Server) generateToken(req *TokenInfo) (authToken oauth2.TokenInfo, err error) {
	tgr := &oauth2.TokenGenerateRequest{
		ClientID:     req.ClientID,
		Scope:        req.Scope,
		RedirectURI:  req.RedirectURI,
		Request:      req.Ctx.Request,
		ClientSecret: req.ClientSecret,
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
		Refresh:      req.RefreshToken,
		//AccessTokenExp: req.AccessTokenExp,
		UserID: req.UserID,
	}
	res, err := s.Manager.GenerateAccessToken(req.Ctx.Request.Context(), oauth2.GrantType(req.GrantType), tgr)
	//这里可以根据不同的grant模式返回不同的错误(比如client模式可以返回更详尽的错误类型)
	if err != nil {
		switch err {
		case errors.ErrInvalidAuthorizeCode, errors.ErrInvalidCodeChallenge, errors.ErrMissingCodeChallenge:
			return nil, errors.ErrInvalidGrant
		case errors.ErrInvalidClient:
			return nil, errors.ErrInvalidClient
		default:
			return nil, err
		}
	}
	return res, nil
}

// func (s *Server) refreshToken(req *TokenInfo) (authToken oauth2.TokenInfo, err error) {
// 	s.Manager.Re
// rti, err := s.Manager.LoadRefreshToken(ctx, tgr.Refresh)
// 			if err != nil {
// 				if err == errors.ErrInvalidRefreshToken || err == errors.ErrExpiredRefreshToken {
// 					return nil, errors.ErrInvalidGrant
// 				}
// 				return nil, err
// 			}
// }
