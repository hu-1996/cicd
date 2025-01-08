package handler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	jobexec "cicd-server/job_exec"
	"cicd-server/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func Events(ctx context.Context, c *app.RequestContext) {
	var event types.Event
	if err := c.BindAndValidate(&event); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	jobexec.AddEvent(&event)

	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}

func Log(ctx context.Context, c *app.RequestContext) {
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
			if err := os.MkdirAll(filepath.Dir(logPath), os.ModePerm); err != nil {
				hlog.Errorf("mkdir error: %s", err)
				c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
				return
			}
			file, err := os.Create(logPath)
			if err != nil {
				hlog.Errorf("create file error: %s", err)
				c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
				return
			}
			defer file.Close()
			_, err = file.WriteString(log.Log + "\n")
			if err != nil {
				hlog.Errorf("write file error: %s", err)
				c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
				return
			}
			c.JSON(consts.StatusOK, utils.H{"data": "success"})
			return
		} else {
			hlog.Errorf("stat file error: %s", err)
			c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
			return
		}
	}

	file, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		hlog.Errorf("open file error: %s", err)
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	defer file.Close()
	_, err = file.WriteString(log.Log + "\n")
	if err != nil {
		hlog.Errorf("write file error: %s", err)
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}
