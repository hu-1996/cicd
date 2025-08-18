package jobexec

import (
	"bufio"
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cicd-runner/types"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

var (
	jobChan       = make(chan *JobExec, 5)
	serverUrl     string
	name          string
	eventChan     = make(chan *types.Event, 1000)
	logChan       = make(chan *types.Log, 1000)
	mutex         sync.Mutex
	jobCancelFunc = make(map[uint]context.CancelFunc)
	jobMap        = make(map[uint]*JobExec)
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
	CommitId   string
}

type JobExec struct {
	Job       Job
	JobRunner JobRunner
	Git       Git
}

func (j *JobExec) AddJob() {
	jobChan <- j
}

func CancelJob(jobRunnerID uint) {
	hlog.Infof("cancel job: %d", jobRunnerID)
	mutex.Lock()
	defer mutex.Unlock()
	if job, ok := jobMap[jobRunnerID]; ok {
		job.AddLog("Received job interruption request, processing...")
	}
	if cancel, ok := jobCancelFunc[jobRunnerID]; ok {
		cancel()
	}
}

func (j *JobExec) AddLog(log string) {
	if log != "" {
		log = fmt.Sprintf("%s [%s] %s", time.Now().Format("2006/01/02 15:04:05"), name, log)
	}

	hlog.Infof("add log[%d]: %s", j.JobRunner.ID, log)
	logChan <- &types.Log{
		JobRunnerID: j.JobRunner.ID,
		Log:         log,
	}
}

func (j *JobExec) AddEvent(success bool, message string) {
	if message != "" {
		message = fmt.Sprintf("[%s] %s", name, message)
	}
	eventChan <- &types.Event{
		JobRunnerID: j.JobRunner.ID,
		Success:     success,
		Message:     message, //message,
	}
}

func Run(n, su string) {
	hlog.Infof("start job exec")
	name = n
	serverUrl = su

	go handleEvent()
	go handleLog()

	for job := range jobChan {
		job.Exec()
	}
}

func (job *JobExec) Exec() {
	mutex.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	jobCancelFunc[job.JobRunner.ID] = cancel
	jobMap[job.JobRunner.ID] = job
	mutex.Unlock()
	job.RunCommand(ctx)
}

func (job *JobExec) RunCommand(ctx context.Context) {
	if job == nil {
		hlog.Info("job channel closed")
		return
	}
	hlog.Infof("start job: %+v", job)

	var dir string
	if job.Git.ID > 0 {
		if job.Git.CommitId == "" {
			job.AddEvent(false, "commit id is empty")
			job.AddLog("commit id is empty")
			return
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			hlog.Errorf("get home dir error: %s", err)
			job.AddEvent(false, err.Error())
			job.AddLog(err.Error())
			return
		}
		dir = filepath.Join(homeDir, ".cicd-runner", "repos", fmt.Sprintf("%d", job.Git.ID))

		if err = job.GitCloneOrPull(dir); err != nil {
			hlog.Errorf("git clone or pull error: %s", err)
			job.AddEvent(false, err.Error())
			job.AddLog(err.Error())
			return
		}

		job.AddLog(fmt.Sprintf("git clone or pull success, branch: %s, commit id: %s", job.Git.Branch, job.Git.CommitId))
	}

	job.Job.Envs = append(job.Job.Envs, Env{
		Key: "VERSION",
		Val: job.Job.Tag,
	})
	for _, env := range job.Job.Envs {
		job.AddLog(fmt.Sprintf("set env: %s=%s", env.Key, env.Val))
		if err := os.Setenv(env.Key, env.Val); err != nil {
			hlog.Errorf("set env error: %s", err)
			job.AddEvent(false, err.Error())
			job.AddLog(err.Error())
			return
		}
	}

	succeed := true
	defer func() {
		if succeed {
			job.AddEvent(true, "")
			job.AddLog("This step was executed successfully.")
		}

		// 清除环境变量
		for _, env := range job.Job.Envs {
			if err := os.Unsetenv(env.Key); err != nil {
				hlog.Errorf("unset env error: %s", err)
			}
		}

		mutex.Lock()
		delete(jobCancelFunc, job.JobRunner.ID)
		mutex.Unlock()
	}()
	for _, command := range job.JobRunner.Commands {
		select {
		case <-ctx.Done():
			job.AddEvent(false, "job interrupted")
			job.AddLog("job interrupted")
			return
		default:
			succeed = job.command(ctx, dir, command)
			if !succeed {
				return
			}
		}
	}

}

func (job *JobExec) command(ctx context.Context, dir, command string) bool {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	hlog.Infof("run command: %s", command)

	job.AddLog(fmt.Sprintf("%s$ %s", cmp.Or(dir, "~"), command))
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

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			job.AddLog(scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			job.AddLog(scanner.Text())
		}
	}()

	err = cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			hlog.Errorf("exit code: %v", exitErr)
			job.AddEvent(false, fmt.Sprintf("exit code: %d", exitErr.ExitCode()))
			job.AddLog(fmt.Sprintf("exit code: %d", exitErr.ExitCode()))
			return false
		} else {
			hlog.Errorf("run command error: %s", err)
			job.AddEvent(false, err.Error())
			job.AddLog(err.Error())
			return false
		}
	}
	hlog.Info("----------------------------------------")
	return true
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

func (j *JobExec) GitCloneOrPull(dir string) error {
	var err error
	var clone bool

	_, err = os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			clone = true
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if clone {
		if err = j.cloneRepo(dir); err != nil {
			return handleExitError(err, dir)
		}
	} else {
		// 获取远程分支
		if err := j.fetchBranch(dir); err != nil {
			return err
		}

		// 切换到远程分支
		if err = j.checkoutBranch(dir); err != nil {
			return err
		}

		// 拉取远程分支
		if err = j.pullRepo(dir); err != nil {
			return handleExitError(err, dir)
		}
	}

	// 切换到commit id
	if err = j.checkoutCommit(dir); err != nil {
		return err
	}

	hlog.Infof("git clone or pull success, dir: %s", dir)
	return nil
}

func (j *JobExec) cloneRepo(dir string) error {
	var cmd *exec.Cmd
	if j.Git.Username != "" && j.Git.Password != "" {
		// http(s) 认证
		authUrl := strings.Replace(j.Git.Repository, "://", fmt.Sprintf("://%s:%s@", j.Git.Username, j.Git.Password), 1)
		cmd = exec.Command("git", "clone", "-b", j.Git.Branch, "--single-branch", authUrl, dir)
	} else {
		// ssh 或匿名
		cmd = exec.Command("git", "clone", "-b", j.Git.Branch, "--single-branch", j.Git.Repository, dir)
	}

	// 创建管道获取标准输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// 创建管道获取标准错误
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			j.AddLog(scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			j.AddLog(scanner.Text())
		}
	}()

	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (j *JobExec) pullRepo(dir string) error {
	cmd := exec.Command("git", "-C", dir, "pull")

	// 创建管道获取标准输出
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// 创建管道获取标准错误
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			j.AddLog(scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			j.AddLog(scanner.Text())
		}
	}()

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (j *JobExec) fetchBranch(dir string) error {
	cmd := exec.Command("git", "-C", dir, "fetch")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return handleExitError(err, dir)
	}
	j.AddLog(string(out))
	return nil
}

func (j *JobExec) checkoutBranch(dir string) error {
	cmd := exec.Command("git", "-C", dir, "checkout", j.Git.Branch)
	out, err := cmd.CombinedOutput()
	if err != nil {
		hlog.Errorf("checkout branch error: %s", err)
		return handleExitError(err, dir)
	}
	j.AddLog(string(out))
	return nil
}

func (j *JobExec) checkoutCommit(dir string) error {
	cmd := exec.Command("git", "-C", dir, "checkout", j.Git.CommitId)
	out, err := cmd.CombinedOutput()
	if err != nil {
		hlog.Errorf("checkout commit error: %s", err)
		return handleExitError(err, dir)
	}
	j.AddLog(string(out))
	return nil
}

func handleExitError(err error, dir string) error {
	if exitErr, ok := err.(*exec.ExitError); ok {
		hlog.Infof("remove dir: %s", dir)
		if err := os.RemoveAll(dir); err != nil {
			hlog.Errorf("remove dir error: %s", err)
		}

		return fmt.Errorf("exit code: %d, error: %s", exitErr.ExitCode(), string(exitErr.Stderr))
	}
	return err
}
