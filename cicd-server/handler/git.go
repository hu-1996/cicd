package handler

import (
	"context"

	"cicd-server/git"
	"cicd-server/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func TestGit(ctx context.Context, c *app.RequestContext) {
	var req types.TestGitReq
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, utils.H{"error": err.Error()})
		return
	}

	lastCommit, err := git.RepoLastCommit(req.Repository, req.Branch, req.Username, req.Password)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, utils.H{"message": "success", "last_commit": lastCommit})
}
