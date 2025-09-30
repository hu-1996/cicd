package dal

import (
	"cicd-server/types"
	"cicd-server/utils"
	"errors"
	"time"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"size:32" vd:"len($)>0"`
	Password string `json:"password" gorm:"size:64;not null" vd:"len($)>0"`
	Nickname string `json:"nickname" gorm:"size:32"`
}

func (u *User) Format() *types.UserResp {
	var userRoles []UserRole
	if err := DB.Where("user_id = ?", u.ID).Find(&userRoles).Error; err != nil {
		return &types.UserResp{
			ID:       u.ID,
			Username: u.Username,
			Nickname: u.Nickname,
		}
	}

	var roles []Role
	if len(userRoles) > 0 {
		roleIds := lo.Map(userRoles, func(role UserRole, _ int) uint {
			return role.RoleID
		})
		DB.Where("id in ?", roleIds).Find(&roles)
	}
	return &types.UserResp{
		ID:       u.ID,
		Username: u.Username,
		Nickname: u.Nickname,
		Roles: lo.Map(userRoles, func(role UserRole, _ int) uint {
			return role.RoleID
		}),
		CreatedAt: u.CreatedAt.Format(time.DateTime),
		UpdatedAt: u.UpdatedAt.Format(time.DateTime),
		IsAdmin: lo.ContainsBy(roles, func(role Role) bool {
			return role.Name == "admin"
		}),
	}
}

func (u *User) CreateAdmin() error {
	var users []User
	password := utils.Sha256([]byte(u.Password))
	if err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("username = ?", u.Username).Find(&users).Error; err != nil {
			return err
		}
		role, err := CreateAdminRole(tx)
		if err != nil {
			return err
		}
		if len(users) > 0 {
			user := users[0]

			var userRoles UserRole
			if err := tx.Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&userRoles).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return tx.Create(&UserRole{
						UserID: user.ID,
						RoleID: role.ID,
					}).Error
				}
				return err
			}
			return nil
		}
		u.Password = password
		u.Nickname = "admin"
		if err := tx.Create(u).Error; err != nil {
			return err
		}
		return tx.Create(&UserRole{
			UserID: u.ID,
			RoleID: role.ID,
		}).Error
	}); err != nil {
		return err
	}
	return nil
}

type UserRole struct {
	gorm.Model
	UserID uint
	RoleID uint
}
