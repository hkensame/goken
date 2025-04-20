package authmodel

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"github.com/go-oauth2/oauth2/v4"
)

type CodeGener struct {
	NodeNum int64
	Node    *snowflake.Node
}

func MustNewCodeGener(n int64) *CodeGener {
	c := &CodeGener{
		NodeNum: n,
	}
	var err error
	c.Node, err = snowflake.NewNode(n)
	if err != nil {
		panic(err)
	}
	return c
}

func (g *CodeGener) Token(ctx context.Context, data *oauth2.GenerateBasic) (code string, err error) {
	code = g.Node.Generate().Base64()
	return
}

func (r *Auther) Token(ctx context.Context, data *oauth2.GenerateBasic, isGenRefresh bool) (string, string, error) {
	// clentInfo, ok := data.Client.(*AuthClientInfo)
	// if !ok {
	// 	return "", "", errors.New("错误的data数据")
	// }

	token, _, err := r.Jwt.NewToken("aud", data.Client.GetID(), "sub", data.UserID)
	if err != nil {
		return "", "", err
	}
	return token, token, nil
}
