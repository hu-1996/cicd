package types

type Event struct {
	JobRunnerID uint   `path:"job_runner_id" vd:"$>0"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
}

type Log struct {
	JobRunnerID uint   `path:"job_runner_id" vd:"$>0"`
	Log         string `json:"log"`
}
