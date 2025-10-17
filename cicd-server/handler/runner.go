package handler

import (
	"context"
	"errors"

	"cicd-server/dal"
	"cicd-server/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func RegisterRunner(ctx context.Context, c *app.RequestContext) {
	var runner types.RegisterRunnerReq
	if err := c.BindAndValidate(&runner); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	if err := dal.DB.Transaction(func(tx *gorm.DB) error {
		var r dal.Runner
		if err := tx.Last(&r, "name = ?", runner.Name).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				r = dal.Runner{
					Name:     runner.Name,
					Endpoint: runner.Endpoint,
					Status:   dal.Online,
					Message:  "",
					IP:       runner.IP,
				}
				if err := tx.Create(&r).Error; err != nil {
					return err
				}

				for _, label := range runner.Labels {
					if err := tx.Create(&dal.RunnerLabel{
						RunnerID: r.ID,
						Label:    label,
					}).Error; err != nil {
						return err
					}
				}
				return nil
			}
			return err
		}
		if r.Endpoint != runner.Endpoint {
			return errors.New("runner name already exists")
		} else {
			r.Name = runner.Name
			r.Status = dal.Online
			r.IP = runner.IP
			if err := tx.Save(&r).Error; err != nil {
				return err
			}
			if err := tx.Delete(&dal.RunnerLabel{}, "runner_id = ?", r.ID).Error; err != nil {
				return err
			}

			for _, label := range runner.Labels {
				if err := tx.Create(&dal.RunnerLabel{
					RunnerID: r.ID,
					Label:    label,
				}).Error; err != nil {
					return err
				}
			}
			return nil
		}
	}); err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}

func ListRunner(ctx context.Context, c *app.RequestContext) {
	var req types.ListRunnerReq
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var runners []dal.Runner
	var total int64
	if err := dal.DB.
		Order("id desc").
		Scopes(dal.Paginate(req.Page, req.PageSize)).
		Find(&runners).
		Offset(-1).Limit(-1).
		Count(&total).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if len(runners) == 0 {
		c.JSON(consts.StatusOK, utils.H{
			"list":  []types.RunnerResp{},
			"total": 0,
		})
		return
	}
	rs := lo.Map(runners, func(job dal.Runner, _ int) types.RunnerResp {
		return job.Format()
	})

	c.JSON(consts.StatusOK, utils.H{
		"list":  rs,
		"total": total,
	})
}

func EnableRunner(ctx context.Context, c *app.RequestContext) {
	var runner types.PathRunnerReq
	if err := c.BindAndValidate(&runner); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var r dal.Runner
	if err := dal.DB.First(&r, "id = ?", runner.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	r.Enable = !r.Enable
	if r.Status == dal.Offline {
		r.Enable = false
	}

	if err := dal.DB.Save(&r).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}

func SetRunnerBusy(ctx context.Context, c *app.RequestContext) {
	var runner types.PathRunnerReq
	if err := c.BindAndValidate(&runner); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var r dal.Runner
	if err := dal.DB.First(&r, "id = ?", runner.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	r.PipelineID = 0
	r.PipelineName = ""
	if err := dal.DB.Save(&r).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}

func DeleteRunner(ctx context.Context, c *app.RequestContext) {
	var runner types.PathRunnerReq
	if err := c.BindAndValidate(&runner); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	if err := dal.DB.Delete(&dal.Runner{}, "id = ?", runner.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}

func ListRunnerLabel(ctx context.Context, c *app.RequestContext) {
	var runnerLabels []dal.RunnerLabel
	if err := dal.DB.Find(&runnerLabels).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	rs := lo.Map(runnerLabels, func(item dal.RunnerLabel, _ int) string {
		return item.Label
	})

	c.JSON(consts.StatusOK, lo.Uniq(rs))
}
