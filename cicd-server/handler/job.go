package handler

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cicd-server/dal"
	jobexec "cicd-server/job_exec"
	"cicd-server/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func PipelineJobs(ctx context.Context, c *app.RequestContext) {
	var job types.PipelineJobReq
	if err := c.BindAndValidate(&job); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var jobs []dal.Job
	var total int64
	if err := dal.DB.
		Order("id desc").
		Scopes(dal.Paginate(job.Page, job.PageSize)).
		Find(&jobs, "pipeline_id = ?", job.PipelineID).
		Offset(-1).Limit(-1).
		Count(&total).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if len(jobs) == 0 {
		c.JSON(consts.StatusOK, utils.H{
			"list":  []types.JobResp{},
			"total": 0,
		})
		return
	}
	js := lo.Map(jobs, func(job dal.Job, _ int) types.JobResp {
		return job.Format()
	})

	c.JSON(consts.StatusOK, utils.H{
		"list":  js,
		"total": total,
	})
}

func StartJob(ctx context.Context, c *app.RequestContext) {
	var job types.StartJobReq
	if err := c.BindAndValidate(&job); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var pipeline dal.Pipeline
	if err := dal.DB.Last(&pipeline, "id = ?", job.PipelineID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	var steps []dal.Step
	if err := dal.DB.Find(&steps, "pipeline_id = ?", job.PipelineID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if err := dal.DB.Transaction(func(tx *gorm.DB) error {
		mapEnv := make(map[string]string)
		for _, env := range pipeline.Envs {
			mapEnv[env.Key] = env.Val
		}
		for _, env := range job.Envs {
			mapEnv[env.Key] = env.Val
		}

		var envs []dal.Env
		for k, v := range mapEnv {
			envs = append(envs, dal.Env{
				Key: k,
				Val: v,
			})
		}
		job := dal.Job{
			PipelineID: job.PipelineID,
			Envs:       envs,
		}
		if err := tx.Create(&job).Error; err != nil {
			return err
		}

		var newTag string
		switch {
		case strings.Contains(pipeline.TagTemplate, "${COUNT}"):
			newTag = strings.ReplaceAll(pipeline.TagTemplate, "${COUNT}", strconv.Itoa(int(job.ID)))
		case strings.Contains(pipeline.TagTemplate, "${TIMESTAMP}"):
			newTag = strings.ReplaceAll(pipeline.TagTemplate, "${TIMESTAMP}", strconv.FormatInt(time.Now().Unix(), 10))
		case strings.Contains(pipeline.TagTemplate, "${DATETIME}"):
			newTag = strings.ReplaceAll(pipeline.TagTemplate, "${DATETIME}", time.Now().Format("20060102150405"))
		}
		job.Tag = cmp.Or(newTag, pipeline.TagTemplate)

		if err := tx.Save(&job).Error; err != nil {
			return err
		}

		var runners []dal.JobRunner
		for _, step := range steps {
			runner := dal.JobRunner{
				JobID:    job.ID,
				StepID:   step.ID,
				Status:   dal.Pending,
				Trigger:  step.Trigger,
				Commands: step.Commands,
			}
			if err := tx.Create(&runner).Error; err != nil {
				return err
			}

			runners = append(runners, runner)
		}

		// job exec
		if len(runners) > 0 {
			var git dal.Git
			if pipeline.UseGit {
				if err := tx.Last(&git, "pipeline_id = ?", job.PipelineID).Error; err != nil {
					return err
				}
				git.Pull = true
			}
			jobexec.NewJobExec(job, runners[0], git).AddJob()
		}

		return nil
	}); err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}

func StartJobStep(ctx context.Context, c *app.RequestContext) {
	var job types.StartStepReq
	if err := c.BindAndValidate(&job); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var jobRunner dal.JobRunner
	if err := dal.DB.Last(&jobRunner, "id = ?", job.JobRunnerID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	var j dal.Job
	if err := dal.DB.Last(&j, "id = ?", jobRunner.JobID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	var pipeline dal.Pipeline
	if err := dal.DB.Last(&pipeline, "id = ?", j.PipelineID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if err := dal.DB.Transaction(func(tx *gorm.DB) error {
		jobRunner.ID = 0
		jobRunner.Status = dal.Pending
		jobRunner.Message = ""
		jobRunner.AssignRunnerIds = []uint{}
		jobRunner.EventStatus = map[dal.Status]int{}
		jobRunner.Trigger = dal.TriggerManual
		if err := tx.Create(&jobRunner).Error; err != nil {
			return err
		}
		var git dal.Git
		if pipeline.UseGit {
			if err := tx.Last(&git, "pipeline_id = ?", j.PipelineID).Error; err != nil {
				return err
			}
		}
		jobexec.NewJobExec(j, jobRunner, git).AddJob()
		return nil
	}); err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}

func JobRunnerDetail(ctx context.Context, c *app.RequestContext) {
	var runner types.PathJobRunnerReq
	if err := c.BindAndValidate(&runner); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var jobRunner dal.JobRunner
	if err := dal.DB.Last(&jobRunner, "id = ?", runner.JobRunnerID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	var job dal.Job
	if err := dal.DB.Last(&job, "id = ?", jobRunner.JobID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	var pipeline dal.Pipeline
	if err := dal.DB.Last(&pipeline, "id = ?", job.PipelineID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	var steps []dal.Step
	if err := dal.DB.Find(&steps, "pipeline_id = ?", job.PipelineID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	sts := lo.Map(steps, func(step dal.Step, _ int) types.StepResp {
		return step.Format()
	})
	stepBy := lo.Associate(sts, func(step types.StepResp) (uint, types.StepResp) {
		return step.ID, step
	})

	var jobRunners []dal.JobRunner
	var jrs []types.JobRunner
	var jr types.JobRunner
	if err := dal.DB.Find(&jobRunners, "job_id = ? AND step_id = ?", jobRunner.JobID, jobRunner.StepID).Error; err == nil {
		for _, jobRunner := range jobRunners {
			rs := types.JobRunner{
				LastRunnerID: jobRunner.ID,
				LastStatus:   string(jobRunner.Status),
				StartTime:    lo.If(jobRunner.StartTime.IsZero(), "-").Else(jobRunner.StartTime.Format("2006-01-02 15:04:05")),
				EndTime:      lo.If(jobRunner.StartTime.IsZero(), "-").Else(jobRunner.EndTime.Format("2006-01-02 15:04:05")),
				Cost:         jobRunner.EndTime.Sub(jobRunner.StartTime).String(),
				Message:      jobRunner.Message,
			}
			if s, ok := stepBy[jobRunner.StepID]; ok {
				rs.StepName = s.Name
			}

			for _, runnerId := range jobRunner.AssignRunnerIds {
				var runner dal.Runner
				if err := dal.DB.Last(&runner, "id = ?", runnerId).Error; err == nil {
					rs.AssignRunners = append(rs.AssignRunners, runner.Format())
				} else {
					hlog.Errorf("get runner[%d] error: %s", runnerId, err)
				}
			}
			jrs = append(jrs, rs)
			if jobRunner.ID == runner.JobRunnerID {
				jr = rs
			}
		}
	} else {
		hlog.Errorf("get jobRunners error: %s", err)
	}

	resp := types.JobRunnerResp{
		Pipeline:   pipeline.Format(),
		Steps:      sts,
		JobRunners: jrs,
		Job:        job.Format(),
		JobRunner:  jr,
	}

	c.JSON(consts.StatusOK, resp)
}

func JobRunnerLog(ctx context.Context, c *app.RequestContext) {
	var log types.Log
	if err := c.BindAndValidate(&log); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		hlog.Errorf("get home dir error: %s", err)
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	logPath := filepath.Join(homeDir, ".cicd-runner", "logs", fmt.Sprintf("%d.log", log.JobRunnerID))
	_, err = os.Stat(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(consts.StatusInternalServerError, utils.H{"error": "log not found"})
			return
		}
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	file, err := os.ReadFile(logPath)
	if err != nil {
		hlog.Errorf("open file error: %s", err)
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, string(file))
}
