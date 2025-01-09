package utils

import (
	"context"
	"encoding/json"
	"errors"

	"aidanwoods.dev/go-paseto"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

type User struct {
	Id       uint   `json:"id"`
	Username string `json:"username"`
}

func LoginUser(ctx context.Context, c *app.RequestContext) (*User, error) {
	if data, ok := c.Get("paseto"); ok {
		token := data.(paseto.Token)
		claims := token.ClaimsJSON()

		var user User
		err := json.Unmarshal(claims, &user)
		if err != nil {
			hlog.Errorf("unmarshal user: %v", err)
			return nil, err
		}
		return &user, nil
	}
	return nil, errors.New("failed to retrieve user information")
}
