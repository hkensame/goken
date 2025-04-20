package authmodel

import (
	"net/http"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/hkensame/goken/pkg/errors"
)

var (
	ErrInternalFailed error = errors.New("内部错误")
)

func (r *Auther) checkAllowedGrant(clientID string, grant oauth2.GrantType) (bool, error) {
	res := AuthClientInfo{ClientID: clientID}
	if err := r.DB.db.Where("client_id = ?", clientID).First(&res).Error; err != nil {
		r.Logger.Sugar().Errorf("db查询信息失败 err = %v", err)
		return false, ErrInternalFailed
	}
	gl := grantList{}
	if err := gl.UnmarshalJSON(res.GrantTypes); err != nil {
		r.Logger.Sugar().Errorf("格式化失败 err = %v", err)
		return false, ErrInternalFailed
	}
	if gl.HasGrant(grant.String()) {
		return true, nil
	}
	return false, nil
}

func (r *Auther) checkAllowedScope(tgr *oauth2.TokenGenerateRequest) (bool, error) {
	res := AuthClientInfo{ClientID: tgr.ClientID}
	if err := r.DB.db.Where("client_id = ?", tgr.ClientID).First(&res).Error; err != nil {
		r.Logger.Sugar().Errorf("db查询信息失败 err = %v", err)
		return false, ErrInternalFailed
	}
	sl := scopeList{}
	if err := sl.UnmarshalJSON(res.Scope); err != nil {
		r.Logger.Sugar().Errorf("格式化失败 err = %v", err)
		return false, ErrInternalFailed
	}
	if sl.HasScope(tgr.Scope) {
		return true, nil
	}
	return false, nil
}

// 默认把clientInfo放在FormData中
func (r *Auther) extractClientInfo(req *http.Request) (string, string, error) {
	//auth := req.Header.Get("Authorization")
	// if strings.HasPrefix(auth, "Basic ") {
	// 	return server.ClientBasicHandler(req)
	// }
	return server.ClientFormHandler(req)
}

func (r *Auther) authorizeUser(w http.ResponseWriter, req *http.Request) (userID string, err error) {
	req.Header.Get("Authorization")
}
