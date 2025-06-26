package jobexec

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

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

	go handleEvent()
	go handleLog()

	for job := range jobChan {
		if job == nil {
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
				if !succeed {
					continue
				}
				cmd := exec.Command("sh", "-c", command)
				cmd.Dir = dir
				hlog.Infof("run command: %s", command)

				addLog(job.JobRunner.ID, fmt.Sprintf("%s# %s", dir, command))
				// 创建管道获取标准输出
				stdout, err := cmd.StdoutPipe()
				if err != nil {
					hlog.Errorf("cmd stdout pipe error: %s", err)
				}

				// 创建管道获取标准错误
				stderr, err := cmd.StderrPipe()
				if err != nil {
					hlog.Errorf("cmd stderr pipe error: %s", err)
				}

				if err := cmd.Start(); err != nil {
					hlog.Errorf("cmd start error: %s", err)
				}
				var wg sync.WaitGroup
				wg.Add(2)
				go func() {
					defer wg.Done()
					scanner := bufio.NewScanner(stdout)
					for scanner.Scan() {
						addLog(job.JobRunner.ID, fmt.Sprintf("[stdout] %s", scanner.Text()))
					}
				}()
				go func() {
					defer wg.Done()
					scanner := bufio.NewScanner(stderr)
					for scanner.Scan() {
						addLog(job.JobRunner.ID, fmt.Sprintf("[stderr] %s", scanner.Text()))
					}
				}()

				err = cmd.Wait()
				wg.Wait()
				if err != nil {
					if exitErr, ok := err.(*exec.ExitError); ok {
						hlog.Errorf("exit code: %d", exitErr.ExitCode())
						addEvent(job.JobRunner.ID, false, fmt.Sprintf("exit code: %d", exitErr.ExitCode()))
						addLog(job.JobRunner.ID, fmt.Sprintf("exit code: %d", exitErr.ExitCode()))
						succeed = false
					} else {
						hlog.Errorf("run command error: %s", err)
						addEvent(job.JobRunner.ID, false, err.Error())
						addLog(job.JobRunner.ID, err.Error())
						succeed = false
					}
				}
				// output, err := cmd.Output()
				// if err != nil {
				// 	hlog.Errorf("run command error: %s", err)
				// 	addEvent(job.JobRunner.ID, false, err.Error())
				// 	addLog(job.JobRunner.ID, err.Error())
				// 	succeed = false
				// 	continue
				// }
				// addLog(job.JobRunner.ID, fmt.Sprintf("output:\n %s", string(output)))
				// hlog.Infof("run command success: %s", string(output))
				hlog.Info("----------------------------------------")
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

func handleEvent() {
	for event := range eventChan {
		sendEvent(event)
	}
}

func sendEvent(event *types.Event) {
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

func addLog(runnerId uint, log string) {
	if log != "" {
		log = fmt.Sprintf("[%s] %s", name, log)
	}
	logChan <- &types.Log{
		JobRunnerID: runnerId,
		Log:         log,
	}
}

func handleLog() {
	for log := range logChan {
		sendLog(log)
	}
}

func sendLog(log *types.Log) {
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
