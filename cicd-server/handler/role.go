package handler

import (
	"context"

	"cicd-server/dal"
	"cicd-server/types"
	cutils "cicd-server/utils"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/samber/lo"
)

func CreateRole(ctx context.Context, c *app.RequestContext) {
	user, err := cutils.LoginUser(ctx, c)
	if err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}
	if !user.IsAdmin {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "无权限"})
		return
	}

	var req types.RoleReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	role := dal.Role{
		Name: req.Name,
	}

	if err := dal.DB.Create(&role).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, utils.H{"id": role.Format()})
}

func DeleteRole(ctx context.Context, c *app.RequestContext) {
	user, err := cutils.LoginUser(ctx, c)
	if err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}
	if !user.IsAdmin {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "无权限"})
		return
	}

	var req types.RolePathReq
	err = c.BindAndValidate(&req)
	if err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var role dal.Role
	if err := dal.DB.Where("id = ?", req.Id).First(&role).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	if role.Name == "admin" {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "不能删除admin角色"})
		return
	}

	if err := dal.DB.Delete(&role).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, utils.H{"message": "删除成功"})
}

func ListRole(ctx context.Context, c *app.RequestContext) {
	var roles []dal.Role
	if err := dal.DB.Find(&roles).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, utils.H{"list": lo.Map(roles, func(role dal.Role, _ int) *types.Role {
		return role.Format()
	})})
}
