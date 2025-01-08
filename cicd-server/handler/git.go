package handler

import (
	"context"
	"strings"

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

	repoSplit := strings.Split(req.Repository, "/")
	repoName := strings.TrimSuffix(repoSplit[len(repoSplit)-1], ".git")

	commitId, err := git.TestClone(repoName, req.Repository, req.Branch, req.Username, req.Password)
	if err != nil {
		c.JSON(consts.StatusInternalServerError, utils.H{"error": err.Error()})
		return
	}

	c.JSON(consts.StatusOK, utils.H{"commit_id": commitId})
}
