package types

type RegisterRunnerReq struct {
	Name     string   `json:"name" vd:"regexp('^[a-zA-Z0-9_-]+$')"`
	Endpoint string   `json:"endpoint"`
	Labels   []string `json:"labels"`
}

type RunnerResp struct {
	ID        uint     `json:"id"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	Enable    bool     `json:"enable"`
	Busy      bool     `json:"busy"`
	Labels    []string `json:"labels"`
	CreatedAt string   `json:"created_at"`
}

type ListRunnerReq struct {
	Name     string `query:"name"`
	Page     int    `query:"page"`
	PageSize int    `query:"page_size"`
}

type PathRunnerReq struct {
	ID uint `path:"id" vd:"$>0"`
}