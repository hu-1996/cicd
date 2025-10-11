package dal

import (
	"cicd-server/types"

	"gorm.io/gorm"
)

type Runner struct {
	gorm.Model
	Name          string
	Endpoint      string
	Status        RunnerStatus
	Message       string
	PipelineID    uint
	PipelineName  string
	Enable        bool `gorm:"default:1"`
	StageID       uint
	StageParallel bool
}

type RunnerStatus string

const (
	Online  RunnerStatus = "online"
	Offline RunnerStatus = "offline"
)

func (r *Runner) Format() types.RunnerResp {
	resp := types.RunnerResp{
		ID:           r.ID,
		Name:         r.Name,
		PipelineID:   r.PipelineID,
		PipelineName: r.PipelineName,
		Status:       string(r.Status),
		Enable:       r.Enable,
		CreatedAt:    r.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	var labels []RunnerLabel
	if err := DB.Find(&labels, "runner_id = ?", r.ID).Error; err == nil {
		for _, label := range labels {
			resp.Labels = append(resp.Labels, label.Label)
		}
	}
	return resp
}
