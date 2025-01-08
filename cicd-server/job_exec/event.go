package jobexec

import (
	"cicd-server/dal"
	"cicd-server/types"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/samber/lo"
)

var eventChan = make(chan *types.Event, 10000)

func AddEvent(event *types.Event) {
	eventChan <- event
}

func StartEventProcess() {
	for {
		select {
		case event := <-eventChan:
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
				dal.DB.Model(&dal.Runner{}).Where("id = ?", runnerId).Update("busy", false)
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
				StartNextStep(jobRunner.ID)
			}
		}
	}
}

func StartNextStep(jobRunnerID uint) {
	var jobRunner dal.JobRunner
	if err := dal.DB.Last(&jobRunner, "id = ?", jobRunnerID).Error; err != nil {
		return
	}

	var job dal.Job
	if err := dal.DB.Last(&job, "id = ?", jobRunner.JobID).Error; err != nil {
		return
	}

	var runners []dal.JobRunner
	if err := dal.DB.Find(&runners, "job_id = ? AND status = ? AND id > ?", jobRunner.JobID, dal.Pending, jobRunnerID).Error; err != nil {
		return
	}
	if len(runners) > 0 {
		nextRunner := runners[0]
		if nextRunner.Trigger == dal.TriggerManual {
			return
		} else if jobRunner.Status == dal.Success {
			var git dal.Git
			if err := dal.DB.Last(&git, "pipeline_id = ?", job.PipelineID).Error; err != nil {
				return
			}
			NewJobExec(job, nextRunner, git).AddJob()
		}
	}
}
