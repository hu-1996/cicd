package types

type Role struct {
	Id        uint   `path:"id" json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type RoleReq struct {
	Name string `json:"name"`
}

type RolePathReq struct {
	Id uint `path:"id" vd:"$>0"`
}
