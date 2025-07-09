package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// 判断 URL 是否是 SSH URL
func isSSHURL(url string) bool {
	return len(url) >= 4 && url[:4] == "git@"
}

func RepoLastCommit(repoUrl, branch, username, password string) (string, error) {
	// 构建git命令
	var gitCmd string
	if username != "" && password != "" {
		// 使用用户名密码认证
		repoUrlWithAuth := strings.Replace(repoUrl, "https://", fmt.Sprintf("https://%s:%s@", username, password), 1)
		gitCmd = fmt.Sprintf("git ls-remote %s %s | awk '{print $1}'", repoUrlWithAuth, branch)
	} else if isSSHURL(repoUrl) {
		// 使用SSH认证
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}

		privateKeyFile := filepath.Join(homeDir, ".ssh", "id_rsa")
		hlog.Infof("privateKeyFile: %s", privateKeyFile)
		if _, err := os.Stat(privateKeyFile); err != nil {
			return "", errors.New("private key file not found in path: " + privateKeyFile)
		}
		gitCmd = fmt.Sprintf("GIT_SSH_COMMAND='ssh -i %s' git ls-remote %s %s | awk '{print $1}'", privateKeyFile, repoUrl, branch)
	} else {
		gitCmd = fmt.Sprintf("git ls-remote %s %s | awk '{print $1}'", repoUrl, branch)
	}

	// 执行git命令
	cmd := exec.Command("sh", "-c", gitCmd)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git command failed: %v, output: %s", exitErr.ExitCode(), string(output))
		}
		return "", fmt.Errorf("git command failed: %v, output: %s", err, string(output))
	}

	commitId := strings.TrimSpace(string(output))
	if commitId == "" {
		return "", errors.New("获取到的提交ID为空")
	}

	return commitId, nil
}
