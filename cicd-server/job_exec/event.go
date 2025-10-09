package jobexec

import (
	"errors"
	"time"

	"cicd-server/dal"
	"cicd-server/types"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

var eventChan = make(chan *types.Event, 10000)

func AddEvent(event *types.Event) {
	eventChan <- event
}

func StartEventProcess() {
	for event := range eventChan {
		var jobRunner dal.JobRunner
		if err := dal.DB.Last(&jobRunner, "id = ?", event.JobRunnerID).Error; err != nil {
			hlog.Errorf("get job runner error: %s", err)
			continue
		}

		// 没有分配runner，直接添加到事件队列
		if len(jobRunner.AssignRunnerIds) == 0 {
			AddEvent(event)
			continue
		}

		eventStatus := jobRunner.EventStatus
		for _, runnerId := range jobRunner.AssignRunnerIds {
			if err := dal.DB.Model(&dal.Runner{}).Where("id = ?", runnerId).Updates(map[string]interface{}{"pipeline_id": 0, "pipeline_name": ""}).Error; err != nil {
				hlog.Errorf("update runner error: %s", err)
			}
		}
		eventStatus[lo.Ternary(event.Success, dal.Success, dal.Failed)]++
		updateColumns := map[string]interface{}{
			"event_status": eventStatus,
		}

		var sum int
		for _, count := range jobRunner.EventStatus {
			sum += count
		}

		if sum == len(jobRunner.AssignRunnerIds) {
			if c, ok := jobRunner.EventStatus[dal.Success]; ok && c == len(jobRunner.AssignRunnerIds) {
				jobRunner.Status = dal.Success
			} else if c > 0 {
				jobRunner.Status = dal.PartialSuccess
			} else {
				jobRunner.Status = dal.Failed
			}
			updateColumns["status"] = jobRunner.Status
		}

		if event.Message != "" {
			updateColumns["message"] = jobRunner.Message + event.Message + "; "
		}
		updateColumns["end_time"] = time.Now()

		if err := dal.DB.Model(&dal.JobRunner{}).Where("id = ?", jobRunner.ID).Updates(updateColumns).Error; err != nil {
			hlog.Errorf("update job runner error: %s", err)
			continue
		}

		if sum == len(jobRunner.AssignRunnerIds) {
			// 判断是否有下一步
			if !StartNextStep(jobRunner.ID) {
				StartOtherStep(jobRunner.AssignRunnerIds)
			}
		}
	}
}

func StartNextStep(jobRunnerID uint) bool {
	var jobRunner dal.JobRunner
	if err := dal.DB.Last(&jobRunner, "id = ?", jobRunnerID).Error; err != nil {
		return false
	}
	if jobRunner.Status != dal.Success {
		return false
	}

	var job dal.Job
	if err := dal.DB.Last(&job, "id = ?", jobRunner.JobID).Error; err != nil {
		return false
	}

	tempJobRunnerID := jobRunnerID
	var oldRunner dal.JobRunner
	if err := dal.DB.First(&oldRunner, "job_id = ? AND step_id = ?", jobRunner.JobID, jobRunner.StepID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}
	tempJobRunnerID = oldRunner.ID

	var runners []dal.JobRunner
	if err := dal.DB.Find(&runners, "job_id = ? AND status = ? AND id > ?", job.ID, dal.Pending, tempJobRunnerID).Error; err != nil {
		return false
	}
	if len(runners) > 0 {
		nextRunner := runners[0]
		if nextRunner.Trigger == dal.TriggerManual {
			return false
		} else {
			var git dal.Git
			if err := dal.DB.Last(&git, "pipeline_id = ?", job.PipelineID).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					return false
				}
			} else {
				git.CommitID = job.CommitID
				git.Branch = job.Branch
			}

			nextRunner.Status = dal.Queueing
			if err := dal.DB.Save(&nextRunner).Error; err != nil {
				hlog.Errorf("update job runner error: %s", err)
				return false
			}
			NewJobExec(job, nextRunner, git).AddJob()
			return true
		}
	}

	return false
}

func StartOtherStep(assignRunnerIds []uint) {
	if len(assignRunnerIds) == 0 {
		return
	}

	var jobRunners []dal.JobRunner
	if err := dal.DB.Find(&jobRunners, "status = ?", dal.Queueing).Error; err != nil {
		return
	}

	if len(jobRunners) == 0 {
		return
	}

	var runnerLabels []dal.RunnerLabel
	if err := dal.DB.Find(&runnerLabels, "runner_id IN (?)", assignRunnerIds).Error; err != nil {
		return
	}

	if len(runnerLabels) == 0 {
		return
	}

	rlm := make(map[string]uint)
	for _, label := range runnerLabels {
		rlm[label.Label] = label.RunnerID
	}

	var steps []dal.Step
	if err := dal.DB.Find(&steps, "id IN (?)", lo.Map(jobRunners, func(item dal.JobRunner, _ int) uint {
		return item.StepID
	})).Error; err != nil {
		hlog.Errorf("get steps error: %s", err)
		return
	}

	if len(steps) == 0 {
		return
	}

	stepMap := lo.Associate(steps, func(item dal.Step) (uint, dal.Step) {
		return item.ID, item
	})

	usedAssignRunnerIds := make(map[uint]struct{}, len(assignRunnerIds))
	for _, jobRunner := range jobRunners {
		step, ok := stepMap[jobRunner.StepID]
		if !ok {
			continue
		}
		if runnerId, ok := rlm[step.RunnerLabelMatch]; ok {
			if _, ok := usedAssignRunnerIds[runnerId]; ok {
				continue
			}

			var job dal.Job
			if err := dal.DB.Last(&job, "id = ?", jobRunner.JobID).Error; err != nil {
				return
			}

			var git dal.Git
			if err := dal.DB.Last(&git, "pipeline_id = ?", job.PipelineID).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					return
				}
			} else {
				git.CommitID = job.CommitID
				git.Branch = job.Branch
			}
			usedAssignRunnerIds[runnerId] = struct{}{}
			NewJobExec(job, jobRunner, git).AddJob()
			break
		}
	}
}
