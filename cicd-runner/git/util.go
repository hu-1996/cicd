package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func GitCloneOrPull(dir, repoUrl, branch, username, password string) (string, error) {
	var err error
	var clone bool

	_, err = os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			clone = true
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	var cmd *exec.Cmd
	if clone {
		// 构造 clone 命令
		if username != "" && password != "" {
			// http(s) 认证
			authUrl := strings.Replace(repoUrl, "://", fmt.Sprintf("://%s:%s@", username, password), 1)
			cmd = exec.Command("git", "clone", "-b", branch, "--single-branch", authUrl, dir)
		} else {
			// ssh 或匿名
			cmd = exec.Command("git", "clone", "-b", branch, "--single-branch", repoUrl, dir)
		}

	} else {
		// pull
		cmd = exec.Command("git", "-C", dir, "pull", "origin", branch)
	}
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
			hlog.Infof("[stdout] %s", scanner.Text())
		}
	}()
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			hlog.Infof("[stderr] %s", scanner.Text())
		}
	}()

	err = cmd.Wait()
	wg.Wait()
	if err != nil {
		return "", err
	}

	hlog.Infof("git clone or pull success, dir: %s", dir)

	// 获取 commit id
	cmd = exec.Command("git", "-C", dir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
