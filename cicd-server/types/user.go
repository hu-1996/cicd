package types

type UserReq struct {
	ID       uint   `path:"id" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Roles    []uint `json:"roles"`
}

type UpdateUserReq struct {
	ID       uint   `path:"id" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Roles    []uint `json:"roles"`
}

type UserResp struct {
	ID        uint   `path:"id" json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Roles     []uint `json:"roles"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	IsAdmin   bool   `json:"is_admin"`
}

type ListUserReq struct {
	Nickname string `query:"nickname"`
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
}

type LoginUser struct {
	Username string `json:"username" gorm:"size:32;unique" vd:"len($)>0"`
	Password string `json:"password" gorm:"size:64;not null" vd:"len($)>0"`
}

type UserPathReq struct {
	Id uint `path:"id" vd:"$>0"`
}
