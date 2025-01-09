package handler

import (
	"context"
	"errors"
	"time"

	"cicd-server/dal"
	"cicd-server/types"
	cutils "cicd-server/utils"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/paseto"
	"gorm.io/gorm"
)

func Login(ctx context.Context, c *app.RequestContext) {
	var user types.LoginUser
	err := c.BindAndValidate(&user)
	if err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var dbUser dal.User
	if err := dal.DB.Where("username = ?", user.Username).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(consts.StatusBadRequest, utils.H{"error": "账号或密码错误"})
			return
		}
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if dbUser.Password != cutils.Sha256([]byte(user.Password)) {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "账号或密码错误"})
		return
	}

	now := time.Now()
	genTokenFunc := paseto.DefaultGenTokenFunc()
	token, err := genTokenFunc(&paseto.StandardClaims{
		Issuer:    "cicd",
		ExpiredAt: now.Add(24 * time.Hour),
		NotBefore: now,
		IssuedAt:  now,
	}, utils.H{
		"id":       dbUser.ID,
		"username": dbUser.Username,
	}, nil)
	if err != nil {
		hlog.Error("generate token failed")
	}

	c.JSON(consts.StatusOK, utils.H{"token": token})
}

func UserInfo(ctx context.Context, c *app.RequestContext) {
	user, err := cutils.LoginUser(ctx, c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, utils.H{"error": err.Error()})
		return
	}

	var dbUser dal.User
	if err := dal.DB.Where("id = ?", user.Id).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(consts.StatusBadRequest, utils.H{"error": "用户不存在"})
			return
		}
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, dbUser.Format())
}
