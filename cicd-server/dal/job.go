package dal

import (
	"cicd-server/types"
	"sort"

	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	PipelineID uint
	Tag        string
	Envs       Envs `gorm:"type:json"`
}

func (j *Job) Format() types.JobResp {
	var evns types.Envs
	for _, v := range j.Envs {
		evns = append(evns, types.Env{
			Key: v.Key,
			Val: v.Val,
		})
	}

	var jobRunners []JobRunner
	jrs := make(map[uint]types.JobRunner)
	if err := DB.Order("id asc").Find(&jobRunners, "job_id = ?", j.ID).Error; err == nil {
		for _, jobRunner := range jobRunners {
			rs := types.JobRunner{
				LastRunnerID: jobRunner.ID,
				StepID:       jobRunner.StepID,
				LastStatus:   string(jobRunner.Status),
			}
			var step Step
			if err := DB.Last(&step, "id = ?", jobRunner.StepID).Error; err == nil {
				rs.StepName = step.Name
			}
			jrs[jobRunner.StepID] = rs
		}
	}
	rs := lo.Values(jrs)
	sort.Slice(rs, func(i, j int) bool {
		return rs[i].StepID < rs[j].StepID
	})

	return types.JobResp{
		ID:         j.ID,
		Tag:        j.Tag,
		Envs:       evns,
		UpdatedAt:  j.UpdatedAt.Format("2006-01-02 15:04:05"),
		JobRunners: rs,
	}
}
