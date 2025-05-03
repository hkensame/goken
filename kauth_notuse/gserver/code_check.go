package gserver

import (
	"fmt"
	"sort"
	"strings"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/hkensame/goken/pkg/encrypt"
)

/*
	对code /token下的检查
*/

// 检查/token方法下得到的tokeninfo与存入的data是否一致
func (s *Server) CheckValidCodeTokenInfo(cd *authCodeData, tk *TokenInfo) error {
	fmt.Println(cd, tk)
	err := s.checkValidCodeChallenge(cd, tk.CodeVerifier, tk.needPKCE)
	if err != nil {
		return err
	}
	if cd.RedirectURI != tk.RedirectURI {
		return errors.ErrInvalidRedirectURI
	}

	if cd.ClientID != tk.ClientID {
		return errors.ErrInvalidClient
	}

	tk.Scope = strings.ReplaceAll(strings.TrimSpace(strings.ToLower(tk.Scope)), ",", " ")
	tkfields := strings.Fields(tk.Scope)
	cdfields := strings.Fields(cd.Scope)
	sort.Strings(tkfields)
	sort.Strings(cdfields)
	for i, _ := range tkfields {
		if tkfields[i] != cdfields[i] {
			return errors.ErrInvalidScope
		}
	}
	return nil
}

// 检查是否满足code challenge
func (s *Server) checkValidCodeChallenge(cd *authCodeData, cv string, needPKCE bool) error {
	if !needPKCE {
		return nil
	}
	switch cd.CodeChallengeMethod {
	case oauth2.CodeChallengeS256.String():
		ccub := encrypt.HashWithDefault([]byte(cv), encrypt.SHA256)
		if string(ccub) != cd.CodeChallenge {
			return errors.ErrUnauthorizedClient
		}
	// 	// 考虑不允许明文
	// case oauth2.CodeChallengePlain:
	// 	return errors.ErrUnsupportedCodeChallengeMethod
	default:
		return errors.ErrUnsupportedCodeChallengeMethod
	}
	return nil
}
