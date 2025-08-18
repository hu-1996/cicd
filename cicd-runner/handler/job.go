package handler

import (
	"context"

	jobexec "cicd-runner/job_exec"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func StartJob(ctx context.Context, c *app.RequestContext) {
	var job jobexec.JobExec
	if err := c.BindAndValidate(&job); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}
	job.AddJob()

	hlog.Infof("start job success")

	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}

type CancelJobReq struct {
	JobRunnerID uint `path:"job_runner_id"`
}

func CancelJob(ctx context.Context, c *app.RequestContext) {
	var req CancelJobReq
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}
	jobexec.CancelJob(req.JobRunnerID)

	hlog.Infof("cancel job success")
	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}
