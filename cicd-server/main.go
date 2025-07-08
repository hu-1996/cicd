package main

import (
	"context"
	"strings"

	"cicd-server/dal"
	"cicd-server/handler"
	jobexec "cicd-server/job_exec"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/paseto"
)

func main() {
	dal.Init()
	go jobexec.Run()
	go jobexec.StartEventProcess()

	h := server.Default(server.WithHostPorts(":8029"))

	h.LoadHTMLGlob("./web/views/*")

	prefix := "/assets"
	root := "./web/assets"
	fs := &app.FS{Root: root, PathRewrite: getPathRewriter(prefix)}
	h.StaticFS(prefix, fs)

	h.GET("/*path", func(c context.Context, ctx *app.RequestContext) {
		ctx.HTML(200, "index.html", nil)
	})

	h.POST("/api/login", handler.Login)

	h.POST("/api/register_runner", handler.RegisterRunner)

	h.POST("/api/events/:job_runner_id", handler.Events)
	h.POST("/api/logs/:job_runner_id", handler.Log)

	h.Use(mws()...)
	h.GET("/api/userinfo", handler.UserInfo)

	h.GET("/api/list_runner", handler.ListRunner)
	h.PUT("/api/enable_runner/:id", handler.EnableRunner)
	h.PUT("/api/set_runner_busy/:id", handler.SetRunnerBusy)
	h.DELETE("/api/delete_runner/:id", handler.DeleteRunner)
	h.GET("/api/list_runner_label", handler.ListRunnerLabel)

	h.POST("/api/start_job/:pipeline_id", handler.StartJob)
	h.POST("/api/start_job_step/:job_runner_id", handler.StartJobStep)
	h.GET("/api/pipeline_jobs/:pipeline_id", handler.PipelineJobs)
	h.GET("/api/job_runner/:job_runner_id", handler.JobRunnerDetail)
	h.GET("/api/job_runner_log/:job_runner_id", handler.JobRunnerLog)
	h.POST("/api/cancel_job_runner/:job_runner_id", handler.CancelJobRunner)

	h.GET("/api/list_step", handler.ListStep)
	h.GET("/api/step/:id", handler.StepDetail)
	h.POST("/api/create_step", handler.CreateStep)
	h.PUT("/api/update_step/:id", handler.UpdateStep)
	h.DELETE("/api/delete_step/:id", handler.DeleteStep)
	h.POST("/api/sort_step/:pipeline_id", handler.SortStep)

	h.GET("/api/list_pipeline", handler.ListPipeline)
	h.POST("/api/sort_pipeline", handler.SortPipeline)
	h.GET("/api/pipeline/:id", handler.PipelineDetail)
	h.POST("/api/create_pipeline", handler.CreatePipeline)
	h.PUT("/api/update_pipeline/:id", handler.UpdatePipeline)
	h.DELETE("/api/delete_pipeline/:id", handler.DeletePipeline)
	h.POST("/api/copy_pipeline/:id", handler.CopyPipeline)

	h.POST("/api/test_git", handler.TestGit)

	// h.StaticFS("/", &app.FS{Root: "./../cicd-web/dist", GenerateIndexPages: true, IndexNames: []string{"index.html"}})

	h.Spin()
}

func getPathRewriter(prefix string) app.PathRewriteFunc {
	// Cannot have an empty prefix
	if prefix == "" {
		prefix = "/"
	}
	// Prefix always start with a '/' or '*'
	if prefix[0] != '/' {
		prefix = "/" + prefix
	}

	// Is prefix a direct wildcard?
	isStar := prefix == "/*"
	// Is prefix a partial wildcard?
	if strings.Contains(prefix, "*") {
		isStar = true
		prefix = strings.Split(prefix, "*")[0]
		// Fix this later
	}
	prefixLen := len(prefix)
	if prefixLen > 1 && prefix[prefixLen-1:] == "/" {
		// /john/ -> /john
		prefixLen--
		prefix = prefix[:prefixLen]
	}
	return func(ctx *app.RequestContext) []byte {
		path := ctx.Path()
		if len(path) >= prefixLen {
			if isStar && string(path[0:prefixLen]) == prefix {
				path = append(path[0:0], '/')
			} else {
				path = path[prefixLen:]
				if len(path) == 0 || path[len(path)-1] != '/' {
					path = append(path, '/')
				}
			}
		}
		if len(path) > 0 && path[0] != '/' {
			path = append([]byte("/"), path...)
		}
		return path
	}
}

func mws() []app.HandlerFunc {
	return []app.HandlerFunc{
		paseto.New(paseto.WithTokenPrefix("Bearer ")),
	}
}
