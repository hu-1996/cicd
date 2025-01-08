package git

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

func GitCloneOrPull(dir, repoUrl, branch, username, password string) (string, error) {
	var auth *http.BasicAuth
	if username != "" && password != "" {
		auth = &http.BasicAuth{
			Username: username,
			Password: password,
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
		r, err := git.PlainClone(dir, false, &git.CloneOptions{
			URL:      repoUrl,
			Progress: os.Stdout,
			Auth:     auth,
		})
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
		err = w.Pull(&git.PullOptions{
			RemoteName: "origin",
			Auth:       auth,
		})
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
