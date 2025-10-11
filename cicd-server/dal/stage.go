package dal

import (
	"cicd-server/types"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"gorm.io/gorm"
)

type Stage struct {
	gorm.Model
	PipelineID uint
	Name       string
	Parallel   bool
	Sort       int
}

func (s *Stage) Format() types.StageResp {
	var stepsResp []types.StepResp
	var steps []Step
	if err := DB.Order("sort ASC, id ASC").Find(&steps, "pipeline_id = ? AND stage_id = ?", s.PipelineID, s.ID).Error; err == nil {
		for _, step := range steps {
			stepsResp = append(stepsResp, step.Format())
		}
	} else {
		hlog.Errorf("get steps error: %s", err)
	}
	return types.StageResp{
		ID:         s.ID,
		PipelineID: s.PipelineID,
		Name:       s.Name,
		Parallel:   s.Parallel,
		Steps:      stepsResp,
	}
}
