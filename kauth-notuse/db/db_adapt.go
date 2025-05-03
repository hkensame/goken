package db

import (
	"context"
	"time"

	"github.com/go-redis/cache/v9"
	"github.com/hkensame/goken/pkg/common/gormutil"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"gorm.io/gorm"
)

type AuthClientInfo struct {
	Model
	ClientID      string            `gorm:"type:varchar(64);uniqueIndex"`
	UserID        string            `gorm:"type:varchar(64)"`
	ClientSecret  string            `gorm:"type:varchar(64);index"`
	GrantTypes    gormutil.GormList `gorm:"type:varchar(80)"`
	Scope         gormutil.GormList `gorm:"type:text"`
	ResponseTypes gormutil.GormList `gorm:"type:varchar(80)"`
	Public        bool              `gorm:"default:true"`
	RedirectURIs  gormutil.GormList `gorm:"type:text"`
	Audience      gormutil.GormList `gorm:"type:text"`
}

// 所有要实现的接口,包括存储access-token,refresh-token,auth-code,client-credential-info
// 以及撤销token,
var (
	_ oauth2.AccessTokenStorage            = &GormClientStore{}
	_ oauth2.RefreshTokenStorage           = &GormClientStore{}
	_ oauth2.AuthorizeCodeStorage          = &GormClientStore{}
	_ oauth2.ClientCredentialsGrantStorage = &GormClientStore{}
	_ oauth2.CoreStorage                   = &GormClientStore{}
	_ oauth2.TokenRevocationStorage        = &GormClientStore{}
)

type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type GormClientStore struct {
	db    *gorm.DB
	cache *cache.Cache
}

func (g *GormClientStore) TableName() string {
	return "oauth_srv"
}

func (g *GormClientStore) AutoMigrate() error {
	return g.db.AutoMigrate(&AuthClientInfo{})
}

func MustNewGormClientStore(db *gorm.DB) *GormClientStore {
	g := &GormClientStore{db: db}
	if err := g.AutoMigrate(); err != nil {
		panic(err)
	}
	return g
}

func (cs *GormClientStore) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	client := &AuthClientInfo{}
	if err := cs.db.Where("client_id = ?", id).Take(&client).Error; err != nil {
		return nil, err
	}
	return client, nil
}

func (c *AuthClientInfo) GetID() string {
	return c.ClientID
}

func (c *AuthClientInfo) GetHashedSecret() []byte {
	return []byte(c.ClientSecret)
}

func (c *AuthClientInfo) GetRedirectURIs() []string {
	return c.RedirectURIs
}

func (c *AuthClientInfo) GetGrantTypes() fosite.Arguments {
	return fosite.Arguments(c.GrantTypes)
}

func (c *AuthClientInfo) GetResponseTypes() fosite.Arguments {
	return fosite.Arguments(c.ResponseTypes)
}

func (c *AuthClientInfo) GetScopes() fosite.Arguments {
	return fosite.Arguments(c.Scope)
}

func (c *AuthClientInfo) IsPublic() bool {
	return c.Public
}

func (c *AuthClientInfo) GetAudience() fosite.Arguments {
	return fosite.Arguments(c.Audience)
}

var _ fosite.Client = &AuthClientInfo{}
