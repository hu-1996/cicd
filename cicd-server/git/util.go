package git

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// 判断 URL 是否是 SSH URL
func isSSHURL(url string) bool {
	return len(url) >= 4 && url[:4] == "git@"
}

func TestConnect(repoUrl, username, password string) error {
	var auth transport.AuthMethod
	if username != "" && password != "" {
		auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	} else if isSSHURL(repoUrl) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		privateKeyFile := filepath.Join(homeDir, ".ssh", "id_rsa")
		_, err = os.Stat(privateKeyFile)
		if err == nil {
			auth, _ = ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
		}
	}

	opts := git.ListOptions{}
	if auth != nil {
		opts.Auth = auth
	}
	_, err := git.NewRemote(nil, &config.RemoteConfig{
		URLs: []string{repoUrl},
	}).List(&opts)
	if err != nil {
		return err
	}

	return nil
}
