package authmodel

import (
	"context"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/hkensame/goken/pkg/common/gormutil"
	"gorm.io/gorm"
)

type AuthClientInfo struct {
	Model
	ClientID      string            `gorm:"type:varchar(64);uniqueIndex"`
	UserID        string            `gorm:"type:varchar(64)"`
	ClientSecret  string            `gorm:"type:varchar(64);index"`
	Domain        string            `gorm:"type:varchar(256)"`
	GrantTypes    gormutil.GormList `gorm:"type:varchar(80)"`
	Scope         gormutil.GormList `gorm:"type:text"`
	ResponseTypes gormutil.GormList `gorm:"type:varchar(80)"`
	Public        bool              `gorm:"default:true"`
	RedirectURIs  gormutil.GormList `gorm:"type:text"`
}

type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type GormClientStore struct {
	db *gorm.DB
}

func (g *GormClientStore) TableName() string {
	return "oauth_srv"
}

func MustNewGormClientStore(db *gorm.DB) *GormClientStore {
	return &GormClientStore{db: db}
}

// 获取客户端信息
func (cs *GormClientStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	client := &AuthClientInfo{}
	if err := cs.db.Where("client_id = ?", id).Take(&client).Error; err != nil {
		return nil, err
	}

	return client, nil
}

func (cs *GormClientStore) InsertOne() {}

func (c *AuthClientInfo) GetID() string {
	return c.ClientID
}

func (c *AuthClientInfo) GetSecret() string {
	return c.ClientSecret
}

func (c *AuthClientInfo) GetDomain() string {
	return c.Domain
}

func (c *AuthClientInfo) IsPublic() bool {
	return c.Public
}

func (c *AuthClientInfo) GetUserID() string {
	return c.UserID
}

func (c *AuthClientInfo) GetResponseType() []string {
	return c.ResponseTypes
}

func (c *AuthClientInfo) GetRedirectURIs() []string {
	return c.RedirectURIs
}
