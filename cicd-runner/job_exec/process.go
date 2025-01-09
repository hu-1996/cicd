package jobexec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"cicd-runner/git"
	"cicd-runner/types"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var (
	jobChan   = make(chan *JobExec, 5)
	serverUrl string
	name      string
	eventChan = make(chan *types.Event, 1000)
	logChan   = make(chan *types.Log, 1000)
)

type Job struct {
	ID   uint
	Tag  string
	Envs Envs `gorm:"type:json"`
}

type Envs []Env

type Env struct {
	Key string
	Val string
}

type JobRunner struct {
	ID       uint
	Commands []string
}

type Git struct {
	ID         uint
	Repository string
	Branch     string
	Username   string
	Password   string
	Pull       bool
}

type JobExec struct {
	Job       Job
	JobRunner JobRunner
	Git       Git
}

func (j *JobExec) AddJob() {
	jobChan <- j
}

func Run(n, su string) {
	hlog.Infof("start job exec")
	name = n
	serverUrl = su

	go sendEvent()
	go sendLog()

	for {
		select {
		case job, ok := <-jobChan:
			if !ok {
				hlog.Info("job channel closed")
				return
			}
			var succeed bool
			hlog.Infof("start job: %+v", job)

			var dir string
			if job.Git.ID != 0 {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					hlog.Errorf("get home dir error: %s", err)
					addEvent(job.JobRunner.ID, false, err.Error())
					addLog(job.JobRunner.ID, err.Error())
					continue
				}
				dir = filepath.Join(homeDir, ".cicd-runner", "repos", fmt.Sprintf("%d", job.Git.ID))
				if job.Git.Pull {
					commitId, err := git.GitCloneOrPull(dir, job.Git.Repository, job.Git.Branch, job.Git.Username, job.Git.Password)
					if err != nil {
						hlog.Errorf("git clone or pull error: %s", err)
						addEvent(job.JobRunner.ID, false, err.Error())
						addLog(job.JobRunner.ID, err.Error())
						continue
					}
					addLog(job.JobRunner.ID, fmt.Sprintf("git clone or pull success, branch: %s, commit id: %s", job.Git.Branch, commitId))
				}
			}

			job.Job.Envs = append(job.Job.Envs, Env{
				Key: "VERSION",
				Val: job.Job.Tag,
			})
			for _, env := range job.Job.Envs {
				addLog(job.JobRunner.ID, fmt.Sprintf("set env: %s=%s", env.Key, env.Val))
				if err := os.Setenv(env.Key, env.Val); err != nil {
					hlog.Errorf("set env error: %s", err)
					addEvent(job.JobRunner.ID, false, err.Error())
					addLog(job.JobRunner.ID, err.Error())
					continue
				}
				succeed = true
			}

			if succeed {
				for _, command := range job.JobRunner.Commands {
					cmd := exec.Command("sh", "-c", command)
					cmd.Dir = dir
					hlog.Infof("run command: %s", command)

					addLog(job.JobRunner.ID, fmt.Sprintf("%s# %s", dir, command))
					output, err := cmd.Output()
					if err != nil {
						hlog.Errorf("run command error: %s", err)
						addEvent(job.JobRunner.ID, false, err.Error())
						addLog(job.JobRunner.ID, err.Error())
						succeed = false
						continue
					}
					addLog(job.JobRunner.ID, fmt.Sprintf("output:\n %s", string(output)))
					hlog.Infof("run command success: %s", string(output))
				}
			}

			if succeed {
				addEvent(job.JobRunner.ID, true, "")
				addLog(job.JobRunner.ID, "This step was executed successfully.")
			}

			// 清除环境变量
			for _, env := range job.Job.Envs {
				if err := os.Unsetenv(env.Key); err != nil {
					hlog.Errorf("unset env error: %s", err)
				}
			}
		}
	}
}

func addEvent(runnerId uint, success bool, message string) {
	if message != "" {
		message = fmt.Sprintf("[%s] %s", name, message)
	}
	eventChan <- &types.Event{
		JobRunnerID: runnerId,
		Success:     success,
		Message:     message, //message,
	}
}

func sendEvent() {
	for event := range eventChan {
		client := &http.Client{}
		jsonBytes, _ := json.Marshal(event)
		httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/events/%d", serverUrl, event.JobRunnerID), bytes.NewReader(jsonBytes))
		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(httpReq)
		if err != nil {
			hlog.Warnf("send event error: %s", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			hlog.Warnf("send event failed, status code: %d", resp.StatusCode)
		}
		hlog.Info("send event success")
	}
}

func addLog(runnerId uint, log string) {
	if log != "" {
		log = fmt.Sprintf("[%s] %s", name, log)
	}
	logChan <- &types.Log{
		JobRunnerID: runnerId,
		Log:         log,
	}
}

func sendLog() {
	for log := range logChan {
		client := &http.Client{}
		jsonBytes, _ := json.Marshal(log)
		httpReq, _ := http.NewRequest("POST", fmt.Sprintf("%s/logs/%d", serverUrl, log.JobRunnerID), bytes.NewReader(jsonBytes))
		httpReq.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(httpReq)
		if err != nil {
			hlog.Warnf("send log error: %s", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				hlog.Warnf("send log error: %s", err)
			}

			hlog.Warnf("send log failed, status code: %d, body: %s", resp.StatusCode, string(body))
		}
		hlog.Info("send log success")
	}
}
