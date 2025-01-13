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
var peroid = time.Second * 10

type JobExec struct {
	Job       dal.Job
	JobRunner dal.JobRunner
	Git       dal.Git
}

func NewJobExec(job dal.Job, jobRunner dal.JobRunner, git dal.Git) *JobExec {
	return &JobExec{
		Job:       job,
		JobRunner: jobRunner,
		Git:       git,
	}
}

func (j *JobExec) AddJob() {
	jobChan <- j
}

func Run() {
	for {
		select {
		case job := <-jobChan:
			for {
				hlog.Infof("start job: %+v", job)
				var j dal.Job
				if err := dal.DB.Last(&j, "id = ?", job.Job.ID).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						hlog.Infof("job[%d] not found", job.Job.ID)
						break
					}
					hlog.Errorf("get job[%d] error: %s", job.Job.ID, err)
					dal.DB.Model(&dal.Job{}).Where("id = ?", job.Job.ID).Update("status", dal.Failed)
					break
				}

				var s dal.Step
				if err := dal.DB.Last(&s, "id = ?", job.JobRunner.StepID).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						hlog.Infof("step[%d] not found", job.JobRunner.StepID)
						break
					}
					hlog.Errorf("get step[%d] error: %s", job.JobRunner.StepID, err)
					updateJobRunner(job.JobRunner.ID, dal.Failed, err.Error(), nil, nil, nil)
					break
				}
				updateJobRunner(job.JobRunner.ID, dal.Assigning, "", nil, nil, nil)

				runners, err := matchRunners(s.RunnerLabelMatch, s.MultipleRunnerExec)
				if err != nil {
					hlog.Errorf("detect idle runners error: %s", err)
					updateJobRunner(job.JobRunner.ID, dal.Failed, err.Error(), nil, nil, nil)
					break
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
							if err := sendJob(runner, *job); err != nil {
								hlog.Errorf("send job error: %s", err)
								if status == "" {
									status = dal.Failed
								} else if status == dal.Success {
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
						updateJobRunner(job.JobRunner.ID, status, message, runnerIds, &start, nil)
					} else {
						// 发送到runner
						if err := sendJob(runners[0], *job); err != nil {
							hlog.Errorf("send job error: %s", err)
							runnerIds := dal.AssignRunnerIds{runners[0].ID}
							updateJobRunner(job.JobRunner.ID, dal.Failed, err.Error(), runnerIds, nil, nil)
							break
						}

						// if step success
						updateJobRunner(job.JobRunner.ID, dal.Running, "", dal.AssignRunnerIds{runners[0].ID}, &start, nil)
					}
					break
				}

				if !s.MultipleRunnerExec {
					hlog.Infof("no idle runner match label: %s, wait for %v to retry", s.RunnerLabelMatch, peroid)
					time.Sleep(peroid)
				}
			}
		}
	}
}

func matchRunners(labelMatch string, multipleRunnerExec bool) ([]dal.Runner, error) {
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

	var runners []dal.Runner
	db := dal.DB.Where("status = ? AND enable = ? AND id IN (?)", dal.Online, true, runnerIds)
	if !multipleRunnerExec {
		db = db.Where("busy = ?", false)
	}
	if err := db.Find(&runners).Error; err != nil {
		return nil, err
	}
	if len(runners) == 0 {
		return nil, fmt.Errorf("no available runner: %s", labelMatch)
	}

	return runners, nil
}

func sendJob(runner dal.Runner, job JobExec) error {
	client := &http.Client{}
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
	runner.Busy = true
	dal.DB.Save(&runner)
	hlog.Info("send job success")
	return nil
}

func updateJobRunner(jobRunnerId uint, status dal.Status, message string, runnerIds dal.AssignRunnerIds, startTime *time.Time, endTime *time.Time) {
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
	dal.DB.Model(&dal.JobRunner{}).Where("id = ?", jobRunnerId).Updates(up)
}
