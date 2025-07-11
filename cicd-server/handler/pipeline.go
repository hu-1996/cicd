package handler

import (
	"context"
	"errors"
	"sort"

	"cicd-server/dal"
	"cicd-server/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func ListPipeline(ctx context.Context, c *app.RequestContext) {
	var pipelines []dal.Pipeline
	if err := dal.DB.Order("sort ASC, id ASC").Find(&pipelines).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
	}

	groupBy := lo.GroupBy(pipelines, func(item dal.Pipeline) string {
		return item.GroupName
	})

	var resp []types.PipelineGroupResp
	for k, v := range groupBy {
		resp = append(resp, types.PipelineGroupResp{
			GroupName: k,
			Pipelines: lo.Map(v, func(item dal.Pipeline, _ int) types.PipelineResp {
				return item.ListFormat()
			}),
		})
	}

	sort.Slice(resp, func(i, j int) bool {
		return resp[i].GroupName < resp[j].GroupName
	})

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
		GroupName:   pipeline.GroupName,
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

	var maxSortPipeline dal.Pipeline
	maxSort := p.Sort
	if pipeline.GroupName != p.GroupName {
		if err := dal.DB.Model(&dal.Pipeline{}).Where("group_name = ?", pipeline.GroupName).Order("sort DESC").Last(&maxSortPipeline).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				maxSort = 0
			} else {
				c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
				return
			}
		} else {
			maxSort = maxSortPipeline.Sort + 1
		}
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
		p.GroupName = pipeline.GroupName
		p.Sort = maxSort
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

func SortPipeline(ctx context.Context, c *app.RequestContext) {
	var pipeline types.SortPipelineReq
	if err := c.BindAndValidate(&pipeline); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	if err := dal.DB.Transaction(func(tx *gorm.DB) error {
		for i, id := range pipeline.PipelineIDs {
			if err := tx.Model(&dal.Pipeline{}).Where("id = ?", id).Update("sort", i).Error; err != nil {
				c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
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

func CopyPipeline(ctx context.Context, c *app.RequestContext) {
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
		newP := p
		newP.ID = 0
		newP.Name = p.Name + "_copy"
		newP.Sort = p.Sort + 1
		if err := tx.Create(&newP).Error; err != nil {
			return err
		}

		if p.UseGit {
			var git dal.Git
			if err := tx.First(&git, "pipeline_id = ?", p.ID).Error; err != nil {
				return err
			}
			git.ID = 0
			git.PipelineID = newP.ID
			if err := tx.Create(&git).Error; err != nil {
				return err
			}
		}

		var steps []*dal.Step
		if err := tx.Find(&steps, "pipeline_id = ?", p.ID).Error; err != nil {
			return err
		}
		for _, step := range steps {
			step.ID = 0
			step.PipelineID = newP.ID
		}

		if err := tx.Create(&steps).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, utils.H{"data": "success"})
}
