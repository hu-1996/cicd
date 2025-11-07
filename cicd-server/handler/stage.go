package handler

import (
	"cicd-server/dal"
	"cicd-server/types"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gorm.io/gorm"
)

func StageDetail(ctx context.Context, c *app.RequestContext) {
	var stage types.PathStageReq
	if err := c.BindAndValidate(&stage); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var s dal.Stage
	if err := dal.DB.First(&s, "id = ?", stage.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, s.Format())
}

func CreateStage(ctx context.Context, c *app.RequestContext) {
	var stage types.CreateStageReq
	if err := c.BindAndValidate(&stage); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var count int64
	if err := dal.DB.Model(&dal.Stage{}).Where("name = ? AND pipeline_id = ?", stage.Name, stage.PipelineID).Count(&count).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "stage name already exists"})
		return
	}

	var s dal.Stage
	s.PipelineID = stage.PipelineID
	s.Name = stage.Name
	s.Parallel = stage.Parallel
	if err := dal.DB.Create(&s).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, s.Format())
}

func UpdateStage(ctx context.Context, c *app.RequestContext) {
	var stage types.UpdateStageReq
	if err := c.BindAndValidate(&stage); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var count int64
	if err := dal.DB.Model(&dal.Stage{}).Where("name = ? AND pipeline_id = ? AND id != ?", stage.Name, stage.PipelineID, stage.ID).Count(&count).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "stage name already exists"})
		return
	}

	var s dal.Stage
	if err := dal.DB.First(&s, "id = ?", stage.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	s.PipelineID = stage.PipelineID
	s.Name = stage.Name
	s.Parallel = stage.Parallel
	if err := dal.DB.Save(&s).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, s.Format())
}

func DeleteStage(ctx context.Context, c *app.RequestContext) {
	var stage types.PathStageReq
	if err := c.BindAndValidate(&stage); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var s dal.Stage
	if err := dal.DB.First(&s, "id = ?", stage.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	dal.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&s).Error; err != nil {
			return err
		}
		if err := tx.Delete(&dal.Step{}, "stage_id = ?", stage.ID).Error; err != nil {
			return err
		}
		return nil
	})
	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}
