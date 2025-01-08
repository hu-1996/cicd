package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// 判断 URL 是否是 SSH URL
func isSSHURL(url string) bool {
	return len(url) >= 4 && url[:4] == "git@"
}

func GitCloneOrPull(dir, repoUrl, branch, username, password string) (string, error) {
	var auth transport.AuthMethod
	if username != "" && password != "" {
		auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	} else if isSSHURL(repoUrl) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		privateKeyFile := filepath.Join(homeDir, ".ssh", "id_rsa")
		_, err = os.Stat(privateKeyFile)
		if err == nil {
			auth, _ = ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
		}
	}

	var clone bool
	_, err := os.Stat(dir)
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

	var repo *git.Repository
	if clone {
		opts := git.CloneOptions{
			URL:           repoUrl,
			Progress:      os.Stdout,
			ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
			SingleBranch:  true,
		}
		if auth != nil {
			opts.Auth = auth
		}
		r, err := git.PlainClone(dir, false, &opts)
		if err != nil {
			return "", err
		}
		repo = r
	} else {
		r, err := git.PlainOpen(dir)
		if err != nil {
			return "", err
		}
		w, err := r.Worktree()
		if err != nil {
			return "", err
		}
		opts := git.PullOptions{
			RemoteName:    "origin",
			ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
			SingleBranch:  true,
		}
		if auth != nil {
			opts.Auth = auth
		}
		err = w.Pull(&opts)
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return "", err
		}
		repo = r
	}

	head, err := repo.Head()
	if err != nil {
		return "", err
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return "", err
	}

	return commit.ID().String(), nil
}
