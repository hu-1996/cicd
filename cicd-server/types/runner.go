package types

type RegisterRunnerReq struct {
	Name     string   `json:"name" vd:"regexp('^[a-zA-Z0-9_-]+$')"`
	Endpoint string   `json:"endpoint"`
	Labels   []string `json:"labels"`
	IP       string   `json:"ip"`
}

type RunnerResp struct {
	ID           uint     `json:"id"`
	Name         string   `json:"name"`
	Status       string   `json:"status"`
	PipelineID   uint     `json:"pipeline_id"`
	PipelineName string   `json:"pipeline_name"`
	Enable       bool     `json:"enable"`
	Labels       []string `json:"labels"`
	CreatedAt    string   `json:"created_at"`
	IP           string   `json:"ip"`
}

type ListRunnerReq struct {
	Name     string `query:"name"`
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
}

type PathRunnerReq struct {
	ID uint `path:"id" vd:"$>0"`
}
