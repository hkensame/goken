package db

import (
	"context"

	"github.com/go-redis/cache/v9"
	"github.com/ory/fosite"
)

func (g *GormClientStore) CreateAccessTokenSession(ctx context.Context, signature string, request fosite.Requester) (err error) {
	if err := g.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   request.GetID(),
		Value: signature,
		//TTL: request.,
	}); err != nil {
		return fosite.ErrServerError
	}
	return nil
}

func (g *GormClientStore) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (request fosite.Requester, err error) {

}

func (g *GormClientStore) DeleteAccessTokenSession(ctx context.Context, signature string) (err error) {

}
