package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"

	"cicd-runner/handler"
	jobexec "cicd-runner/job_exec"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/spf13/cobra"
)

var (
	name, runnerUrl, serverUrl string
	labels                     []string
)

func main() {
	// createUser()
	cmd := &cobra.Command{
		Use:   "cicd-runner",
		Short: "Start cicd-runner",
		Long:  "Start cicd-runner",
		Run: func(cmd *cobra.Command, args []string) {
			registerRunner(name, runnerUrl, serverUrl, labels)

			go jobexec.Run(name, serverUrl)
			h := server.Default(server.WithHostPorts(":5913"))
			h.POST("/start_job", handler.StartJob)

			h.Spin()
		},
	}

	cmd.PersistentFlags().StringVarP(&name, "name", "n", "cicd-runner", "runner name")
	cmd.PersistentFlags().StringVarP(&runnerUrl, "runnerUrl", "r", "http://localhost:5913", "runner server url")
	cmd.PersistentFlags().StringVarP(&serverUrl, "serverUrl", "s", "http://localhost:5912", "server url")
	cmd.PersistentFlags().StringSliceVarP(&labels, "labels", "l", []string{}, "runner labels")

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

func registerRunner(name, runnerUrl, serverUrl string, labels []string) {
	req := RegisterRunnerReq{
		Name:     name,
		Endpoint: runnerUrl,
		Labels:   labels,
	}

	client := &http.Client{}
	jsonBytes, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", serverUrl+"/register_runner", bytes.NewReader(jsonBytes))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalf("register runner failed, status code: %d", resp.StatusCode)
		return
	}
	log.Println("register runner success")
}

type RegisterRunnerReq struct {
	Name     string   `json:"name" vd:"regexp('^[a-zA-Z0-9_-]+$')"`
	Endpoint string   `json:"endpoint"`
	Labels   []string `json:"labels"`
}

func createUser() {
	if runtime.GOOS != "linux" {
		log.Println("only support linux system with useradd command")
		return
	}
	group := "cicd-runner"
	username := "cicd-runner"
	cmd := exec.Command("useradd", "-m", "-g", group, username)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil { // create userexec.Command("useradd", "-m", "-g", "cicd-runner", "cicd-runner").Run()
		hlog.Fatal(err)
		return
	}
	log.Println("create user success, setting uid and gid")
	// 获取用户的 UID 和 GID
	cmd = exec.Command("id", "-u", username)
	output, err := cmd.Output()
	if err != nil {
		hlog.Fatal(err)
	}
	uid := string(output[:len(output)-1]) // 去掉末尾的换行符

	cmd = exec.Command("id", "-g", username)
	output, err = cmd.Output()
	if err != nil {
		hlog.Fatal(err)
	}
	gid := string(output[:len(output)-1])

	hlog.Info("uid:", uid)
	hlog.Info("gid:", gid)

	// 设置进程的 UID 和 GID
	uidint, err := strconv.Atoi(uid)
	if err != nil {
		hlog.Fatal(err)
	}
	gidint, err := strconv.Atoi(gid)
	if err != nil {
		hlog.Fatal(err)
	}
	err = syscall.Setuid(uidint)
	if err != nil {
		hlog.Fatal(err)
	}
	err = syscall.Setgid(gidint)
	if err != nil {
		hlog.Fatal(err)
	}

	hlog.Info(string(output))
}
