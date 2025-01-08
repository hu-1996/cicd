package types

type CreatePipelineReq struct {
	Name        string `json:"name" vd:"regexp('^[a-zA-Z0-9_-]+$')"`
	TagTemplate string `json:"tag_template"`
	Envs        Envs   `json:"envs"`
	UseGit      bool   `json:"use_git"`
	Repository  string `json:"repository"`
	Branch      string `json:"branch"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type Envs []Env

type Env struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type UpdatePipelineReq struct {
	ID          uint   `path:"id" vd:"$>0"`
	Name        string `json:"name" vd:"regexp('^[a-zA-Z0-9_-]+$')"`
	TagTemplate string `json:"tag_template"`
	Envs        Envs   `json:"envs"`
	UseGit      bool   `json:"use_git"`
	Repository  string `json:"repository"`
	Branch      string `json:"branch"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

type PathPipelineReq struct {
	ID uint `path:"id" vd:"$>0"`
}

type PipelineResp struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	TagTemplate  string     `json:"tag_template"`
	Envs         Envs       `json:"envs"`
	LastUpdateAt string     `json:"last_update_at"`
	LastTag      string     `json:"last_tag"`
	Current      int        `json:"current"`
	Steps        []StepResp `json:"steps"`
	UseGit       bool       `json:"use_git"`
	Repository   string     `json:"repository"`
	Branch       string     `json:"branch"`
	Username     string     `json:"username"`
	Password     string     `json:"password"`
}
