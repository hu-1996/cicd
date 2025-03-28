package handler

import (
	"context"

	"cicd-server/dal"
	"cicd-server/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gorm.io/gorm"
)

func ListPipeline(ctx context.Context, c *app.RequestContext) {
	var pipelines []dal.Pipeline
	if err := dal.DB.Find(&pipelines).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
	}

	var resp []types.PipelineResp
	for _, v := range pipelines {
		resp = append(resp, v.Format())
	}

	c.JSON(consts.StatusOK, resp)
}

func PipelineDetail(ctx context.Context, c *app.RequestContext) {
	var pipeline types.PathPipelineReq
	if err := c.BindAndValidate(&pipeline); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var p dal.Pipeline
	if err := dal.DB.First(&p, "id = ?", pipeline.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, p.Format())
}

func CreatePipeline(ctx context.Context, c *app.RequestContext) {
	var pipeline types.CreatePipelineReq
	if err := c.BindAndValidate(&pipeline); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var count int64
	if err := dal.DB.Model(&dal.Pipeline{}).Where("name = ?", pipeline.Name).Count(&count).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "pipeline name already exists"})
		return
	}

	p := dal.Pipeline{
		Name:        pipeline.Name,
		TagTemplate: pipeline.TagTemplate,
		UseGit:      pipeline.UseGit,
	}
	var envs []dal.Env
	for _, v := range pipeline.Envs {
		envs = append(envs, dal.Env{
			Key: v.Key,
			Val: v.Val,
		})
	}
	p.Envs = envs

	if err := dal.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&p).Error; err != nil {
			return err
		}
		if pipeline.UseGit {
			git := dal.Git{
				PipelineID: p.ID,
				Repository: pipeline.Repository,
				Branch:     pipeline.Branch,
				Username:   pipeline.Username,
				Password:   pipeline.Password,
			}
			if err := tx.Create(&git).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, p.Format())
}

func DeletePipeline(ctx context.Context, c *app.RequestContext) {
	var pipeline types.PathPipelineReq
	if err := c.BindAndValidate(&pipeline); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var p dal.Pipeline
	if err := dal.DB.First(&p, "id = ?", pipeline.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if err := dal.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&p).Error; err != nil {
			return err
		}
		if p.UseGit {
			if err := tx.Delete(&dal.Git{}, "pipeline_id = ?", p.ID).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}

func UpdatePipeline(ctx context.Context, c *app.RequestContext) {
	var pipeline types.UpdatePipelineReq
	if err := c.BindAndValidate(&pipeline); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	var count int64
	if err := dal.DB.Model(&dal.Pipeline{}).Where("name = ? AND id != ?", pipeline.Name, pipeline.ID).Count(&count).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(consts.StatusBadRequest, utils.H{"error": "pipeline name already exists"})
		return
	}

	var p dal.Pipeline
	if err := dal.DB.First(&p, "id = ?", pipeline.ID).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	if err := dal.DB.Transaction(func(tx *gorm.DB) error {
		if p.UseGit {
			if err := tx.Delete(&dal.Git{}, "pipeline_id = ?", p.ID).Error; err != nil {
				return err
			}
		}
		p.Name = pipeline.Name
		p.TagTemplate = pipeline.TagTemplate
		p.UseGit = pipeline.UseGit
		var envs []dal.Env
		for _, v := range pipeline.Envs {
			envs = append(envs, dal.Env{
				Key: v.Key,
				Val: v.Val,
			})
		}
		p.Envs = envs
		if err := tx.Save(&p).Error; err != nil {
			return err
		}

		if pipeline.UseGit {
			git := dal.Git{
				PipelineID: p.ID,
				Repository: pipeline.Repository,
				Branch:     pipeline.Branch,
				Username:   pipeline.Username,
				Password:   pipeline.Password,
			}
			if err := tx.Create(&git).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, p.Format())
}
