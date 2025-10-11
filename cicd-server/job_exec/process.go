package jobexec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"cicd-server/dal"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

var jobChan = make(chan *JobExec, 10000)

type JobExec struct {
	Job           dal.Job
	Git           dal.Git
	JobRunner     dal.JobRunner
	AllJobRunners []dal.JobRunner
}

func NewJobExec(job dal.Job, jobRunners []dal.JobRunner, git dal.Git) *JobExec {
	return &JobExec{
		Job:           job,
		Git:           git,
		AllJobRunners: jobRunners,
	}
}

func (j *JobExec) AddJob() {
	jobChan <- j
}

func (j *JobExec) UpdateJobRunner(jobRunner dal.JobRunner, status dal.Status, message string, runnerIds dal.AssignRunnerIds, startTime *time.Time, endTime *time.Time) {
	up := map[string]interface{}{"status": status, "message": message}
	if len(runnerIds) > 0 {
		up["assign_runner_ids"] = runnerIds
	}
	if startTime != nil {
		up["start_time"] = startTime
	}
	if endTime != nil {
		up["end_time"] = endTime
	}
	dal.DB.Model(&dal.JobRunner{}).Where("id = ?", jobRunner.ID).Updates(up)
}

func Run() {
	for job := range jobChan {
		for _, jr := range job.AllJobRunners {
			hlog.Infof("start job runner: %+v", jr)

			if job.Git.ID > 0 {
				if job.Git.CommitID == "" {
					job.UpdateJobRunner(jr, dal.Failed, "git commit id is empty", nil, nil, nil)
					continue
				}
			}

			// 检查任务是否被取消
			var jobRunner dal.JobRunner
			if err := dal.DB.Last(&jobRunner, "id = ?", jr.ID).Error; err != nil {
				hlog.Errorf("get job runner[%d] error: %s", jr.ID, err)
				continue
			}

			if jobRunner.Status != dal.Queueing {
				hlog.Infof("job runner[%d] status is not queueing, skip", jr.ID)
				continue
			}

			var s dal.Step
			if err := dal.DB.Last(&s, "id = ?", jr.StepID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					hlog.Infof("step[%d] not found", jr.StepID)
					continue
				}
				hlog.Errorf("get step[%d] error: %s", jr.StepID, err)
				job.UpdateJobRunner(jr, dal.Failed, err.Error(), nil, nil, nil)
				continue
			}

			runners, err := matchRunners(s.RunnerLabelMatch)
			if err != nil {
				hlog.Errorf("detect idle runners error: %s", err)
				job.UpdateJobRunner(jr, dal.Failed, err.Error(), nil, nil, nil)
				continue
			}
			start := time.Now()
			if len(runners) > 0 {
				if s.MultipleRunnerExec {
					var runnerIds dal.AssignRunnerIds
					var status dal.Status
					var message string
					for _, runner := range runners {
						runnerIds = append(runnerIds, runner.ID)
						// 发送到runner
						if err := sendJob(runner, *job, jr); err != nil {
							hlog.Errorf("send job error: %s", err)
							switch status {
							case "":
								status = dal.Failed
							case dal.Success:
								status = dal.PartialRunning
							}
							message += fmt.Sprintf("send job to runner[%s] error: %s; ", runner.Name, err)
							continue
						}

						// if step success
						if status == dal.Failed || status == dal.PartialRunning {
							status = dal.PartialRunning
						} else {
							status = dal.Running
						}
					}
					job.UpdateJobRunner(jr, status, message, runnerIds, &start, nil)
				} else {
					for _, runner := range runners {
						if runner.PipelineID == 0 || (runner.StageID == jr.StageID && runner.StageParallel) {
							// 发送到runner
							runnerIds := dal.AssignRunnerIds{runner.ID}
							if err := sendJob(runner, *job, jr); err != nil {
								hlog.Errorf("send job error: %s", err)
								job.UpdateJobRunner(jr, dal.Failed, err.Error(), runnerIds, nil, nil)
								break
							}

							// if step success
							job.UpdateJobRunner(jr, dal.Running, "", runnerIds, &start, nil)
							break
						}
					}
				}
			}
		}
	}
}

func matchRunners(labelMatch string) ([]*dal.Runner, error) {
	var runnerLabels []dal.RunnerLabel
	if err := dal.DB.Find(&runnerLabels, "label = ?", labelMatch).Error; err != nil {
		return nil, err
	}

	runnerIds := lo.Map(runnerLabels, func(item dal.RunnerLabel, index int) uint {
		return item.RunnerID
	})

	if len(runnerIds) == 0 {
		return nil, fmt.Errorf("no runner match label: %s", labelMatch)
	}

	var runners []*dal.Runner
	if err := dal.DB.Where("status = ? AND enable = ? AND id IN (?)", dal.Online, true, runnerIds).Find(&runners).Error; err != nil {
		return nil, err
	}
	if len(runners) == 0 {
		return nil, fmt.Errorf("no available runner: %s", labelMatch)
	}

	return runners, nil
}

func sendJob(runner *dal.Runner, job JobExec, jobRunner dal.JobRunner) error {
	client := &http.Client{}
	job.JobRunner = jobRunner
	jsonBytes, _ := json.Marshal(job)
	httpReq, _ := http.NewRequest("POST", runner.Endpoint+"/start_job", bytes.NewReader(jsonBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(httpReq)
	if err != nil {
		if opErr, ok := err.(*net.OpError); ok {
			if sysErr, ok := opErr.Err.(*net.OpError); ok && sysErr.Op == "dial" {
				runner.Status = dal.Offline
				dal.DB.Save(&runner)
			}
		}
		// 或者通过字符串匹配简单判断
		if strings.Contains(err.Error(), "connection refused") {
			runner.Status = dal.Offline
			dal.DB.Save(&runner)
		}
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		if resp.StatusCode == 404 {
			runner.Status = dal.Offline
			dal.DB.Save(&runner)
		}
		return fmt.Errorf("send job failed, status code: %d", resp.StatusCode)
	}

	runner.PipelineID = job.Job.PipelineID
	var pipeline dal.Pipeline
	if err := dal.DB.Last(&pipeline, "id = ?", job.Job.PipelineID).Error; err != nil {
		return fmt.Errorf("get pipeline[%d] error: %s", job.Job.PipelineID, err)
	}

	runner.PipelineName = pipeline.Name
	runner.StageID = jobRunner.StageID
	runner.StageParallel = jobRunner.Parallel
	dal.DB.Save(&runner)
	hlog.Info("send job success")
	return nil
}
