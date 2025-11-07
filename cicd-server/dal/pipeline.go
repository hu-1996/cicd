package dal

import (
	"database/sql/driver"
	"encoding/json"
	"sort"

	"cicd-server/types"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Pipeline struct {
	gorm.Model
	Name        string
	GroupName   string
	TagTemplate string
	Envs        Envs `gorm:"type:json"`
	UseGit      bool `gorm:"default:0"`
	Sort        int  `gorm:"default:0"`
}

type Envs []Env

type Env struct {
	Key string
	Val string
}

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (j *Envs) Scan(value interface{}) error {
	val := make(Envs, 0)
	if err := json.Unmarshal(value.([]byte), &val); err != nil {
		return err
	}
	*j = val
	return nil
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (j Envs) Value() (driver.Value, error) {
	if len(j) == 0 {
		return json.Marshal(Envs{})
	}
	return json.Marshal(j)
}

func (p *Pipeline) Format() types.PipelineResp {
	var evns types.Envs
	for _, v := range p.Envs {
		evns = append(evns, types.Env{
			Key: v.Key,
			Val: v.Val,
		})
	}

	pipeline := types.PipelineResp{
		ID:           p.ID,
		Name:         p.Name,
		GroupName:    p.GroupName,
		TagTemplate:  p.TagTemplate,
		Envs:         evns,
		LastUpdateAt: p.UpdatedAt.Format("2006-01-02 15:04:05"),
		LastTag:      p.TagTemplate,
		UseGit:       p.UseGit,
		Sort:         p.Sort,
	}

	var pipelineRoles []PipelineRole
	DB.Find(&pipelineRoles, "pipeline_id = ?", p.ID)
	if len(pipelineRoles) > 0 {
		pipeline.Roles = lo.Map(pipelineRoles, func(v PipelineRole, _ int) uint { return v.RoleID })
	}

	var git Git
	if err := DB.Last(&git, "pipeline_id = ?", p.ID).Error; err == nil {
		pipeline.UseGit = true
		pipeline.Repository = git.Repository
		pipeline.Branch = git.Branch
		pipeline.Username = git.Username
		pipeline.Password = git.Password
	}

	var job Job
	if err := DB.Last(&job, "pipeline_id = ?", p.ID).Error; err == nil {
		pipeline.LastTag = job.Tag
	}

	var stagesAndSteps []types.StageAndStep
	var steps []Step
	if err := DB.Order("sort ASC, id ASC").Find(&steps, "pipeline_id = ? AND (stage_id = 0 OR stage_id IS NULL)", p.ID).Error; err == nil {
		for _, step := range steps {
			stagesAndSteps = append(stagesAndSteps, types.StageAndStep{
				ID:        step.ID,
				Name:      step.Name,
				Sort:      step.Sort,
				CreatedAt: step.CreatedAt,
				Type:      types.StageAndStepTypeStep,
			})
		}
	} else {
		hlog.Errorf("get steps error: %s", err)
	}

	var stages []Stage
	if err := DB.Order("sort ASC, id ASC").Find(&stages, "pipeline_id = ?", p.ID).Error; err == nil {
		for _, stage := range stages {
			st := stage.Format()
			stagesAndSteps = append(stagesAndSteps, types.StageAndStep{
				ID:        stage.ID,
				Name:      stage.Name,
				Sort:      stage.Sort,
				CreatedAt: stage.CreatedAt,
				Type:      types.StageAndStepTypeStage,
				Children: lo.Map(st.Steps, func(step types.StepResp, _ int) types.StageAndStep {
					return types.StageAndStep{ID: step.ID, Name: step.Name, Sort: step.Sort, CreatedAt: step.CreatedAt, Type: types.StageAndStepTypeStep}
				}),
			})
		}
	} else {
		hlog.Errorf("get stages error: %s", err)
	}

	sort.Slice(stagesAndSteps, func(i, j int) bool {
		if stagesAndSteps[i].Sort != stagesAndSteps[j].Sort {
			return stagesAndSteps[i].Sort < stagesAndSteps[j].Sort
		}
		return stagesAndSteps[i].CreatedAt.Before(stagesAndSteps[j].CreatedAt)
	})
	pipeline.StagesAndSteps = stagesAndSteps

	return pipeline
}

func (p *Pipeline) ListFormat() types.PipelineResp {
	var evns types.Envs
	for _, v := range p.Envs {
		evns = append(evns, types.Env{
			Key: v.Key,
			Val: v.Val,
		})
	}

	pipeline := types.PipelineResp{
		ID:           p.ID,
		Name:         p.Name,
		GroupName:    p.GroupName,
		TagTemplate:  p.TagTemplate,
		Envs:         evns,
		LastUpdateAt: p.UpdatedAt.Format("2006-01-02 15:04:05"),
		LastTag:      p.TagTemplate,
		UseGit:       p.UseGit,
		Sort:         p.Sort,
	}

	var pipelineRoles []PipelineRole
	DB.Find(&pipelineRoles, "pipeline_id = ?", p.ID)
	if len(pipelineRoles) > 0 {
		pipeline.Roles = lo.Map(pipelineRoles, func(v PipelineRole, _ int) uint { return v.RoleID })
	}

	var job Job
	if err := DB.Last(&job, "pipeline_id = ?", p.ID).Error; err == nil {
		pipeline.LastTag = job.Tag
	}

	var sortStepIds []uint
	var jobRunners []JobRunner
	if err := DB.Order("id asc").Find(&jobRunners, "job_id = ?", job.ID).Error; err == nil {
		for _, jobRunner := range jobRunners {
			sortStepIds = append(sortStepIds, jobRunner.StepID)
		}
	}
	sortStepIds = lo.Uniq(sortStepIds)

	var stepsResp []types.StepResp
	var steps []Step
	if err := DB.Order("sort ASC, id ASC").Find(&steps, "pipeline_id = ?", p.ID).Error; err == nil {
		if len(sortStepIds) > 0 {
			for _, id := range sortStepIds {
				for _, step := range steps {
					if step.ID == id {
						stepsResp = append(stepsResp, step.Format())
					}
				}
			}
		} else {
			for _, step := range steps {
				stepsResp = append(stepsResp, step.Format())
			}
		}
	} else {
		hlog.Errorf("get steps error: %s", err)
	}
	pipeline.Steps = stepsResp

	return pipeline
}

type PipelineRole struct {
	gorm.Model
	PipelineID uint
	RoleID     uint
}
