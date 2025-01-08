package handler

import (
	"cicd-server/dal"
	"cicd-server/types"
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func ListStep(ctx context.Context, c *app.RequestContext) {
	var steps []dal.Step
	if err := dal.DB.Find(&steps).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
	}
	c.JSON(consts.StatusOK, utils.H{"data": steps})
}

func StepDetail(ctx context.Context, c *app.RequestContext) {
	var step types.PathStepReq
	if err := c.BindAndValidate(&step); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var p dal.Step
	if err := dal.DB.First(&p, "id = ?", step.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, p.Format())
}

func CreateStep(ctx context.Context, c *app.RequestContext) {
	var step types.CreateStepReq
	if err := c.BindAndValidate(&step); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var count int64
	if err := dal.DB.Model(&dal.Step{}).Where("name = ? AND pipeline_id = ?", step.Name, step.PipelineID).Count(&count).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "step name already exists"})
		return
	}

	var s dal.Step
	s.PipelineID = step.PipelineID
	s.Name = step.Name

	s.Commands = step.Commands
	s.Trigger = dal.Trigger(step.Trigger)
	s.RunnerLabelMatch = step.RunnerLabelMatch
	s.MultipleRunnerExec = step.MultipleRunnerExec
	if err := dal.DB.Create(&s).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, s.Format())
}

func DeleteStep(ctx context.Context, c *app.RequestContext) {
	var step types.PathStepReq
	if err := c.BindAndValidate(&step); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var s dal.Step
	if err := dal.DB.First(&s, "id = ?", step.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	if err := dal.DB.Delete(&s).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}

func UpdateStep(ctx context.Context, c *app.RequestContext) {
	var step types.UpdateStepReq
	if err := c.BindAndValidate(&step); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var count int64
	if err := dal.DB.Model(&dal.Step{}).Where("name = ? AND pipeline_id = ? AND id != ?", step.Name, step.PipelineID, step.ID).Count(&count).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "step name already exists"})
		return
	}

	var s dal.Step
	if err := dal.DB.First(&s, "id = ?", step.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	s.ID = step.ID
	s.PipelineID = step.PipelineID
	s.Name = step.Name

	s.Commands = step.Commands
	s.Trigger = dal.Trigger(step.Trigger)
	s.RunnerLabelMatch = step.RunnerLabelMatch
	s.MultipleRunnerExec = step.MultipleRunnerExec
	if err := dal.DB.Save(&s).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, s.Format())
}
