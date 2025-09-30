package dal

import (
	"cicd-server/types"
	"time"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	Name string `json:"name" gorm:"size:32;" vd:"len($)>0"`
}

func (r *Role) Format() *types.Role {
	return &types.Role{
		Id:        r.ID,
		Name:      r.Name,
		CreatedAt: r.CreatedAt.Format(time.DateTime),
	}
}

func CreateAdminRole(tx *gorm.DB) (*Role, error) {
	var roles []*Role
	if err := tx.Where("name = ?", "admin").Find(&roles).Error; err != nil {
		return nil, err
	}
	if len(roles) > 0 {
		return roles[0], nil
	}
	role := Role{
		Name: "admin",
	}
	err := tx.Create(&role).Error
	return &role, err
}
