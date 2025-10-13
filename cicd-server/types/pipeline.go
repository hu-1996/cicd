package types

import "time"

type CreatePipelineReq struct {
	Name        string `json:"name" vd:"regexp('^[a-zA-Z0-9_-]+$')"`
	GroupName   string `json:"group_name"`
	TagTemplate string `json:"tag_template"`
	Envs        Envs   `json:"envs"`
	UseGit      bool   `json:"use_git"`
	Repository  string `json:"repository"`
	Branch      string `json:"branch"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Sort        int    `json:"sort"`
	Roles       []uint `json:"roles"`
}

type Envs []Env

type Env struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

type UpdatePipelineReq struct {
	ID          uint   `path:"id" vd:"$>0"`
	Name        string `json:"name" vd:"regexp('^[a-zA-Z0-9_-]+$')"`
	GroupName   string `json:"group_name"`
	TagTemplate string `json:"tag_template"`
	Envs        Envs   `json:"envs"`
	UseGit      bool   `json:"use_git"`
	Repository  string `json:"repository"`
	Branch      string `json:"branch"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Sort        int    `json:"sort"`
	Roles       []uint `json:"roles"`
}

type PathPipelineReq struct {
	ID uint `path:"id" vd:"$>0"`
}

type PipelineResp struct {
	ID             uint           `json:"id"`
	Name           string         `json:"name"`
	TagTemplate    string         `json:"tag_template"`
	Envs           Envs           `json:"envs"`
	LastUpdateAt   string         `json:"last_update_at"`
	LastTag        string         `json:"last_tag"`
	Stages         []StageResp    `json:"stages,omitempty"`
	Steps          []StepResp     `json:"steps,omitempty"`
	UseGit         bool           `json:"use_git"`
	Repository     string         `json:"repository"`
	Branch         string         `json:"branch"`
	Username       string         `json:"username"`
	Password       string         `json:"password"`
	GroupName      string         `json:"group_name"`
	Sort           int            `json:"sort"`
	Roles          []uint         `json:"roles"`
	StagesAndSteps []StageAndStep `json:"stages_and_steps"`
}

type StageAndStep struct {
	ID        uint             `json:"id"`
	Name      string           `json:"name"`
	Sort      int              `json:"sort"`
	CreatedAt time.Time        `json:"created_at"`
	Type      StageAndStepType `json:"type"`
	Children  []StageAndStep   `json:"children,omitempty"`
}

type StageAndStepType string

const (
	StageAndStepTypeStage StageAndStepType = "stage"
	StageAndStepTypeStep  StageAndStepType = "step"
)

type PipelineGroupResp struct {
	GroupName string         `json:"group_name"`
	Pipelines []PipelineResp `json:"pipelines"`
}

type SortPipelineReq struct {
	PipelineIDs []uint `json:"pipeline_ids"`
}

type SortStageAndStepReq struct {
	PipelineID uint      `path:"pipeline_id"`
	Stages     []SortReq `json:"stages"`
	Steps      []SortReq `json:"steps"`
}

type SortReq struct {
	ID      uint `json:"id"`
	Sort    int  `json:"sort"`
	StageID uint `json:"stage_id,omitempty"`
}
