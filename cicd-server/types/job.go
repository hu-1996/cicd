package types

type PipelineJobReq struct {
	PipelineID uint `path:"pipeline_id" vd:"$>0"`
	Page       int  `query:"page" vd:"$>0"`
	PageSize   int  `query:"page_size" vd:"$>0"`
}

type StartJobReq struct {
	PipelineID uint `path:"pipeline_id" vd:"$>0"`
	Envs       Envs `json:"envs"`
}

type StartStepReq struct {
	JobRunnerID uint `path:"job_runner_id" vd:"$>0"`
}

type JobResp struct {
	ID         uint        `json:"id"`
	Tag        string      `json:"tag"`
	Envs       Envs        `json:"envs"`
	UpdatedAt  string      `json:"updated_at"`
	JobRunners []JobRunner `json:"job_runners"`
}

type JobRunner struct {
	LastRunnerID  uint         `json:"last_runner_id"`
	StepID        uint         `json:"step_id"`
	StepName      string       `json:"name"`
	StepSort      int          `json:"step_sort"`
	LastStatus    string       `json:"last_status"`
	StartTime     string       `json:"start_time"`
	EndTime       string       `json:"end_time"`
	Cost          string       `json:"cost"`
	Message       string       `json:"message"`
	AssignRunners []RunnerResp `json:"assign_runners"`
}

type PathJobRunnerReq struct {
	JobRunnerID uint `path:"job_runner_id" vd:"$>0"`
}

type JobRunnerResp struct {
	Pipeline   PipelineResp `json:"pipeline"`
	Steps      []StepResp   `json:"steps"`
	JobRunners []JobRunner  `json:"job_runners"`
	Job        JobResp      `json:"job"`
	JobRunner  JobRunner    `json:"job_runner"`
}
