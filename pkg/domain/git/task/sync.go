package task

import (
	"errors"
	"fmt"
	"github.com/MR5356/syncer/pkg/domain/git/config"
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
	"github.com/skeema/knownhosts"
	goSSH "golang.org/x/crypto/ssh"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	isGitUrl           = regexp.MustCompile(`^git@[-\w.:]+:[-\/\w.]+\.git$`)
	isHttpUrl          = regexp.MustCompile(`^(https|http)://[-\w.:]+/[-\/\w.]+\.git$`)
	isTokenizedHttpUrl = regexp.MustCompile(`^(https|http)://[a-zA-Z0-9_]+@[-\w.:]+/[-\/\w.]+\.git$`)
	isBasicHttpUrl     = regexp.MustCompile(`^(https|http)://[a-zA-Z0-9]+:[\w]+@[-\w.:]+/[-\/\w.]+\.git$`)
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
		logrus.Infof("clean %s", dirName)
		_ = os.RemoveAll(dirName)
	}()

	_, err = os.Stat(dirName)
	if err == nil {
		_ = os.RemoveAll(dirName)
	}

	dot = osfs.New(dirName)

	logrus.Infof("clone %s to %s", t.source, dirName)
	repo, err := git.Clone(filesystem.NewStorage(dot, cache.NewObjectLRUDefault()), nil, &git.CloneOptions{
		URL:             repoUrl,
		Mirror:          true,
		Auth:            srcAuth,
		InsecureSkipTLS: true,
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

	logrus.Infof("push to %s", t.destination)
	err = repo.Push(&git.PushOptions{
		RemoteURL:       repoUrl,
		Auth:            destAuth,
		Force:           true,
		InsecureSkipTLS: true,
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
	/**
	支持以下形式：
		1. https://github.com/MR5356/syncer.git
		2. git@github.com:MR5356/syncer.git
		3. https://username:password@github.com/MR5356/syncer.git
		4. https://<token>@github.com/MR5356/syncer.git
		5. https://oauth2:access_token@github.com/MR5356/syncer.git
	*/

	repoUrl = repo
	switch getUrlType(repo) {
	case gitUrlType:
		host, port := parseGitUrl(repo)

		sshKeyScan(host, port)

		if privateKeyFile == "" {
			u, err := user.Current()
			if err == nil {
				privateKeyFile = fmt.Sprintf("%s/.ssh/id_rsa", u.HomeDir)
			}
		}
		_, err = os.Stat(privateKeyFile)
		if err != nil {
			return auth, repoUrl, err
		}
		logrus.Debugf("privateKeyFile: %s, privateKeyPassword: %s", privateKeyFile, privateKeyPassword)
		auth, err = ssh.NewPublicKeysFromFile("git", privateKeyFile, privateKeyPassword)
		return auth, repoUrl, err
	case httpUrlType:
		auth = nil
		return
	case tokenizedHttpUrlType:
		token := strings.ReplaceAll(strings.ReplaceAll(strings.Split(repo, "@")[0], "https://", ""), "http://", "")
		logrus.Infof(token)
		auth = &http.TokenAuth{
			Token: token,
		}
		repoUrl = strings.ReplaceAll(repo, token+"@", "")
		return
	case basicHttpUrlType:
		basicInfo := strings.ReplaceAll(strings.ReplaceAll(strings.Split(repo, "@")[0], "https://", ""), "http://", "")
		fields := strings.Split(basicInfo, ":")
		auth = &http.BasicAuth{
			Username: fields[0],
			Password: fields[1],
		}
		repoUrl = strings.ReplaceAll(repo, basicInfo+"@", "")
		return
	default:
		return nil, "", fmt.Errorf("unsupported repo url: %s", repo)
	}
}

type urlType int

const (
	unknownUrlType urlType = iota
	gitUrlType
	httpUrlType
	tokenizedHttpUrlType
	basicHttpUrlType
)

func getUrlType(url string) (t urlType) {
	if isGitUrl.MatchString(url) {
		t = gitUrlType
	} else if isHttpUrl.MatchString(url) {
		t = httpUrlType
	} else if isTokenizedHttpUrl.MatchString(url) {
		t = tokenizedHttpUrlType
	} else if isBasicHttpUrl.MatchString(url) {
		t = basicHttpUrlType
	} else {
		t = unknownUrlType
	}
	logrus.Debugf("getUrlType: %v", t)
	return t
}

func parseGitUrl(url string) (host string, port int) {
	fields := strings.Split(url, ":")
	port = 22
	if len(fields) == 2 {
		host = strings.Split(fields[0], "@")[1]
	} else if len(fields) == 3 {
		host = strings.Split(fields[0], "@")[1]
		port, _ = strconv.Atoi(fields[1])
	}
	return
}

func sshKeyScan(host string, port int) {
	files, err := getDefaultKnownHostsFiles()
	if err != nil {
		logrus.Warnf("ssh-keyscan getDefaultKnownHostsFiles error: %s", err)
		return
	}
	ks, err := knownhosts.New(files...)
	if err != nil {
		logrus.Warnf("ssh-keyscan knownhosts error: %s", err)
		return
	}
	_, err = goSSH.Dial("tcp", fmt.Sprintf("%s:%d", host, port), &goSSH.ClientConfig{
		HostKeyCallback:   ks.HostKeyCallback(),
		HostKeyAlgorithms: ks.HostKeyAlgorithms(fmt.Sprintf("%s:%d", host, port)),
	})
	if err != nil && strings.Contains(err.Error(), "knownhosts") {
		logrus.Warnf("%s, ssh-keyscan", err)
		u, err := user.Current()
		if err == nil {
			cmd := exec.Command("sh", "-c", fmt.Sprintf("ssh-keyscan -p %d %s >> %s/.ssh/known_hosts", port, host, u.HomeDir))
			logrus.Debugf(cmd.String())
			out, err := cmd.CombinedOutput()
			if err != nil {
				logrus.Warnf("ssh-keyscan error: %s", err)
			}
			logrus.Debugf("ssh-keyscan logs: %s", string(out))
		}
	}
}

func getDefaultKnownHostsFiles() ([]string, error) {
	files := filepath.SplitList(os.Getenv("SSH_KNOWN_HOSTS"))
	if len(files) != 0 {
		return files, nil
	}

	homeDirPath, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return []string{
		filepath.Join(homeDirPath, "/.ssh/known_hosts"),
	}, nil
}
