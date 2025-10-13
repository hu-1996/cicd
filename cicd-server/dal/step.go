package dal

import (
	"database/sql/driver"
	"encoding/json"

	"cicd-server/types"

	"gorm.io/gorm"
)

type Step struct {
	gorm.Model
	PipelineID         uint
	StageID            uint
	Name               string
	Commands           ListString
	Trigger            Trigger
	RunnerLabelMatch   string
	MultipleRunnerExec bool
	Sort               int
}

type ListString []string

func (list ListString) Value() (driver.Value, error) {
	if len(list) == 0 {
		return json.Marshal([]string{})
	}

	return json.Marshal(list)
}

func (list *ListString) Scan(input interface{}) error {
	val := make([]string, 0)
	if err := json.Unmarshal(input.([]byte), &val); err != nil {
		return err
	}
	*list = val
	return nil
}

type Trigger string

const (
	TriggerManual Trigger = "manual"
	TriggerAuto   Trigger = "auto"
)

func (s *Step) Format() types.StepResp {
	step := types.StepResp{
		ID:                 s.ID,
		StageID:            s.StageID,
		Name:               s.Name,
		Commands:           s.Commands,
		Trigger:            string(s.Trigger),
		RunnerLabelMatch:   s.RunnerLabelMatch,
		MultipleRunnerExec: s.MultipleRunnerExec,
		Sort:               s.Sort,
		CreatedAt:          s.CreatedAt,
	}

	var job Job
	var jobRunner JobRunner
	if err := DB.Last(&job, "pipeline_id = ?", s.PipelineID).Error; err == nil {
		if err := DB.Last(&jobRunner, "job_id = ? AND step_id = ?", job.ID, s.ID).Error; err == nil {
			step.LastStatus = string(jobRunner.Status)
			step.LastRunnerID = jobRunner.ID
		}
	}
	step.Parallel = jobRunner.Parallel
	return step
}
