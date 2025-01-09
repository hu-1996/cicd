package types

type User struct {
	ID       uint   `path:"id" json:"id"`
	Username string `json:"username"`
}

type LoginUser struct {
	Username string `json:"username" gorm:"size:32;unique" vd:"len($)>0"`
	Password string `json:"password" gorm:"size:64;not null" vd:"len($)>0"`
}

type UserPathReq struct {
	Id string `path:"id" vd:"len($)>0"`
}
