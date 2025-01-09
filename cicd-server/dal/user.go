package dal

import (
	"cicd-server/types"
	"cicd-server/utils"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"size:32;unique" vd:"len($)>0"`
	Password string `json:"password" gorm:"size:64;not null" vd:"len($)>0"`
}

func (u *User) Format() *types.User {
	return &types.User{
		ID:       u.ID,
		Username: u.Username,
	}
}

func (u *User) Create() error {
	var users []User
	password := utils.Sha256([]byte(u.Password))
	if err := DB.Find(&users).Error; err != nil {
		return err
	}
	if len(users) > 0 {
		user := users[0]
		user.Password = password
		return DB.Save(&user).Error
	} else {
		u.Password = password
		return DB.Create(u).Error
	}
}
