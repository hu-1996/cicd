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

func TestConnect(repoUrl, username, password string) error {
	// 构建git命令
	var gitCmd string
	if username != "" && password != "" {
		// 使用用户名密码认证
		repoUrlWithAuth := strings.Replace(repoUrl, "https://", fmt.Sprintf("https://%s:%s@", username, password), 1)
		gitCmd = fmt.Sprintf("git ls-remote %s", repoUrlWithAuth)
	} else if isSSHURL(repoUrl) {
		// 使用SSH认证
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		privateKeyFile := filepath.Join(homeDir, ".ssh", "id_rsa")
		hlog.Infof("privateKeyFile: %s", privateKeyFile)
		if _, err := os.Stat(privateKeyFile); err != nil {
			return errors.New("private key file not found in path: " + privateKeyFile)
		}
		gitCmd = fmt.Sprintf("GIT_SSH_COMMAND='ssh -i %s' git ls-remote %s", privateKeyFile, repoUrl)
	} else {
		gitCmd = fmt.Sprintf("git ls-remote %s", repoUrl)
	}

	// 执行git命令
	cmd := exec.Command("sh", "-c", gitCmd)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git command failed: %v, output: %s", err, string(output))
	}

	return nil
}
