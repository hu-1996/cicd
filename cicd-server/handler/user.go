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
	"github.com/samber/lo"
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
	var userRoles []dal.UserRole
	if err := dal.DB.Where("user_id = ?", dbUser.ID).Find(&userRoles).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	var roles []dal.Role
	if len(userRoles) > 0 {
		roleIds := lo.Map(userRoles, func(role dal.UserRole, _ int) uint {
			return role.RoleID
		})
		if err := dal.DB.Where("id in ?", roleIds).Find(&roles).Error; err != nil {
			c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
			return
		}
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
		"is_admin": lo.ContainsBy(roles, func(role dal.Role) bool {
			return role.Name == "admin"
		}),
		"nickname": dbUser.Nickname,
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

func CreateUser(ctx context.Context, c *app.RequestContext) {
	var req types.UserReq
	err := c.BindAndValidate(&req)
	if err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var total int64
	if err := dal.DB.Model(&dal.User{}).Where("username = ?", req.Username).Count(&total).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if total > 0 {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "账号已存在"})
		return
	}

	user := dal.User{
		Username: req.Username,
		Nickname: req.Nickname,
		Password: cutils.Sha256([]byte(req.Password)),
	}
	err = dal.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		if len(req.Roles) > 0 {
			for _, role := range req.Roles {
				if err := tx.Create(&dal.UserRole{
					UserID: user.ID,
					RoleID: role,
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, utils.H{"id": user.ID})
}

func UpdateUser(ctx context.Context, c *app.RequestContext) {
	user, err := cutils.LoginUser(ctx, c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, utils.H{"error": err.Error()})
		return
	}

	var req types.UpdateUserReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}
	if !user.IsAdmin && req.ID != user.Id {
		c.JSON(consts.StatusUnauthorized, utils.H{"error": "无权限"})
		return
	}

	var dbUser dal.User
	if err := dal.DB.Where("id = ?", req.ID).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(consts.StatusBadRequest, utils.H{"error": "用户不存在"})
			return
		}
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	dbUser.Nickname = req.Nickname
	err = dal.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&dbUser).Error; err != nil {
			return err
		}

		if err := tx.Delete(&dal.UserRole{}, "user_id = ?", dbUser.ID).Error; err != nil {
			return err
		}

		if len(req.Roles) > 0 {
			for _, role := range req.Roles {
				if err := tx.Create(&dal.UserRole{
					UserID: dbUser.ID,
					RoleID: role,
				}).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, dbUser.Format())
}

func DeleteUser(ctx context.Context, c *app.RequestContext) {
	user, err := cutils.LoginUser(ctx, c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, utils.H{"error": err.Error()})
		return
	}
	if !user.IsAdmin {
		c.JSON(consts.StatusUnauthorized, utils.H{"error": "无权限"})
		return
	}

	var req types.UserPathReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var dbUser dal.User
	if err := dal.DB.Where("id = ?", req.Id).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(consts.StatusBadRequest, utils.H{"error": "用户不存在"})
			return
		}
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	if dbUser.Username == "admin" {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "不能删除admin用户"})
		return
	}

	if err := dal.DB.Delete(&dbUser).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, utils.H{"message": "删除成功"})
}

func ResetPassword(ctx context.Context, c *app.RequestContext) {
	user, err := cutils.LoginUser(ctx, c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, utils.H{"error": err.Error()})
		return
	}

	var resetUser types.UpdateUserReq
	if err := c.BindAndValidate(&resetUser); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}
	if !user.IsAdmin && user.Id != resetUser.ID {
		c.JSON(consts.StatusUnauthorized, utils.H{"error": "无权限"})
		return
	}

	var dbUser dal.User
	if err := dal.DB.Where("id = ?", resetUser.ID).First(&dbUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(consts.StatusBadRequest, utils.H{"error": "用户不存在"})
			return
		}
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if err := dal.DB.Model(&dal.User{}).Where("id = ?", resetUser.ID).Update("password", cutils.Sha256([]byte(resetUser.Password))).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, utils.H{"message": "重置密码成功"})
}

func ListUser(ctx context.Context, c *app.RequestContext) {
	user, err := cutils.LoginUser(ctx, c)
	if err != nil {
		c.JSON(consts.StatusUnauthorized, utils.H{"error": err.Error()})
		return
	}
	if !user.IsAdmin {
		c.JSON(consts.StatusUnauthorized, utils.H{"error": "无权限"})
		return
	}

	var req types.ListUserReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var users []dal.User
	db := dal.DB.Scopes(dal.Paginate(req.Page, req.PageSize))
	if req.Nickname != "" {
		db = db.Where("nickname LIKE ?", "%"+req.Nickname+"%")
	}
	if err := db.Order("id desc").Find(&users).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, utils.H{
		"list": lo.Map(users, func(user dal.User, _ int) *types.UserResp {
			return user.Format()
		}),
		"total": len(users),
	})
}
