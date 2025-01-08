package types

type CreateStepReq struct {
	PipelineID         uint     `json:"pipeline_id" vd:"$>0"`
	Name               string   `json:"name" vd:"regexp('^[a-zA-Z0-9_-]+$')"`
	Commands           []string `json:"commands"`
	Trigger            string   `json:"trigger"`
	RunnerLabelMatch   string   `json:"runner_label_match"`
	MultipleRunnerExec bool     `json:"multiple_runner_exec"`
}

type UpdateStepReq struct {
	ID                 uint     `path:"id" vd:"$>0"`
	PipelineID         uint     `json:"pipeline_id" vd:"$>0"`
	Name               string   `json:"name" vd:"regexp('^[a-zA-Z0-9_-]+$')"`
	Commands           []string `json:"commands"`
	Trigger            string   `json:"trigger"`
	RunnerLabelMatch   string   `json:"runner_label_match"`
	MultipleRunnerExec bool     `json:"multiple_runner_exec"`
}

type PathStepReq struct {
	ID uint `path:"id" vd:"$>0"`
}

type StepResp struct {
	ID                 uint     `json:"id"`
	PipelineID         uint     `json:"pipeline_id"`
	LastRunnerID       uint     `json:"last_runner_id"`
	Name               string   `json:"name"`
	Commands           []string `json:"commands"`
	Trigger            string   `json:"trigger"`
	RunnerLabelMatch   string   `json:"runner_label_match"`
	LastStatus         string   `json:"last_status"`
	MultipleRunnerExec bool     `json:"multiple_runner_exec"`
}
