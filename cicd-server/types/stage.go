package types

type PathStageReq struct {
	ID uint `path:"id"`
}

type StageResp struct {
	ID         uint       `json:"id"`
	PipelineID uint       `json:"pipeline_id"`
	Name       string     `json:"name"`
	Parallel   bool       `json:"parallel"`
	Steps      []StepResp `json:"steps"`
}

type CreateStageReq struct {
	PipelineID uint   `json:"pipeline_id"`
	Name       string `json:"name"`
	Parallel   bool   `json:"parallel"`
}

type UpdateStageReq struct {
	ID         uint   `path:"id"`
	PipelineID uint   `json:"pipeline_id"`
	Name       string `json:"name"`
	Parallel   bool   `json:"parallel"`
}
