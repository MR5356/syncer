package task

import (
	"errors"
	"fmt"
	"github.com/MR5356/syncer/pkg/git/config"
	"github.com/MR5356/syncer/pkg/task"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/sirupsen/logrus"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

type SyncTask struct {
	name        string
	source      string
	destination string

	privateKeyPassword string
	privateKeyFile     string

	ch chan struct{}
}

func NewSyncTask(source, destination, privateKeyFile, privateKeyPassword string, ch chan struct{}) *SyncTask {
	return &SyncTask{
		name:        fmt.Sprintf("%s -> %s", source, destination),
		source:      source,
		destination: destination,

		privateKeyFile:     privateKeyFile,
		privateKeyPassword: privateKeyPassword,

		ch: ch,
	}
}

func GenerateSyncTaskList(cfg *config.Config, ch chan struct{}) (*task.List, error) {
	list := task.NewTaskList()

	for source, dest := range cfg.Repos {
		if destList, ok := dest.([]any); ok {
			if len(destList) == 0 {
				return nil, fmt.Errorf("empty destination for source: %s", source)
			}
			for _, d := range destList {
				if destStr, ok := d.(string); ok {
					logrus.Infof("generate sync task: %s -> %s", source, destStr)
					list.Add(NewSyncTask(source, destStr, cfg.PrivateKeyFile, cfg.PrivateKeyPassword, ch))
				} else {
					return nil, fmt.Errorf("invalid destination type: %T", d)
				}
			}
		} else if destStr, ok := dest.(string); ok {
			if destStr == "" {
				return nil, fmt.Errorf("empty destination for source: %s", source)
			}
			logrus.Infof("generate sync task: %s -> %s", source, destStr)
			list.Add(NewSyncTask(source, destStr, cfg.PrivateKeyFile, cfg.PrivateKeyPassword, ch))
		} else {
			return nil, fmt.Errorf("invalid destination, should be string or []string for source: %s", source)
		}
	}
	return list, nil
}

func (t *SyncTask) Name() string {
	return t.name
}

func (t *SyncTask) Run() error {
	// 源仓库拉取
	srcAuth, repoUrl, err := getAuth(t.source, t.privateKeyFile, t.privateKeyPassword)
	if err != nil {
		return err
	}
	t.source = repoUrl

	var dot billy.Filesystem
	dirName := fmt.Sprintf("/tmp/%s", filepath.Base(t.source))
	defer func() {
		_ = os.RemoveAll(dirName)
	}()

	_, err = os.Stat(dirName)
	if err == nil {
		_ = os.RemoveAll(dirName)
	}

	dot = osfs.New(dirName)

	repo, err := git.Clone(filesystem.NewStorage(dot, cache.NewObjectLRUDefault()), nil, &git.CloneOptions{
		URL:    repoUrl,
		Mirror: true,
		Auth:   srcAuth,
	})
	if err != nil {
		return err
	}

	// 推送到目标仓库
	destAuth, repoUrl, err := getAuth(t.destination, t.privateKeyFile, t.privateKeyPassword)
	if err != nil {
		return err
	}
	t.destination = repoUrl

	err = repo.Push(&git.PushOptions{
		RemoteURL: repoUrl,
		Auth:      destAuth,
		Force:     true,
		RefSpecs: []gitConfig.RefSpec{
			"+refs/heads/*:refs/heads/*",
			"+refs/tags/*:refs/tags/*",
			"+refs/change/*:refs/change/*",
		},
	})

	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		logrus.Warnf("%s is up to date", t.destination)
		return nil
	}
	return err
}

func getAuth(repo, privateKeyFile, privateKeyPassword string) (auth transport.AuthMethod, repoUrl string, err error) {
	repoUrl = repo
	if strings.HasPrefix(repo, "git") {
		if privateKeyFile == "" {
			u, err := user.Current()
			if err != nil {
				return auth, repoUrl, err
			}
			privateKeyFile = fmt.Sprintf("%s/.ssh/id_rsa", u.HomeDir)
		}
		_, err = os.Stat(privateKeyFile)
		if err != nil {
			return auth, repoUrl, err
		}
		auth, err = ssh.NewPublicKeysFromFile("git", privateKeyFile, privateKeyPassword)
	} else if strings.HasPrefix(repo, "http") {
		if strings.Contains(repo, "@") {
			fields := strings.Split(repo, "@")
			userInfo := strings.Join(fields[0:len(fields)-1], "@")
			repoInfo := fields[len(fields)-1]

			if strings.HasPrefix(userInfo, "http://") {
				userInfo = strings.ReplaceAll(userInfo, "http://", "")
				repoUrl = fmt.Sprintf("%s%s", "http://", repoInfo)
			} else if strings.HasPrefix(userInfo, "https://") {
				userInfo = strings.ReplaceAll(userInfo, "https://", "")
				repoUrl = fmt.Sprintf("%s%s", "https://", repoInfo)
			}

			fields = strings.Split(userInfo, ":")
			username := fields[0]
			password := strings.Join(fields[1:], ":")

			auth = &http.BasicAuth{
				Username: username,
				Password: password,
			}
			logrus.Debugf("http auth: username: %s, password: %s, repoUrl: %s", username, password, repoUrl)
		} else {
			auth = nil
		}
	} else {
		auth = nil
	}
	return
}
