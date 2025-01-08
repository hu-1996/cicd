package git

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func TestClone(name, repoUrl, branch, username, password string) (string, error) {
	var auth *http.BasicAuth
	if username != "" && password != "" {
		auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
	}

	dir, err := os.MkdirTemp("", name)
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout,
		Auth:     auth,
	})
	if err != nil {
		return "", err
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
