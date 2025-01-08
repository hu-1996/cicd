package types

type TestGitReq struct {
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}
