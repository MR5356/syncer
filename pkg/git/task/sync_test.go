package task

import (
	"fmt"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"os/user"
	"reflect"
	"testing"
)

func Test_getAuth(t *testing.T) {
	type args struct {
		repo               string
		privateKeyFile     string
		privateKeyPassword string
	}
	tests := []struct {
		name        string
		args        args
		wantAuth    transport.AuthMethod
		wantRepoUrl string
		wantErr     bool
	}{
		{
			name: "test git url",
			args: args{
				repo:               "git@github.com:MR5356/syncer.git",
				privateKeyFile:     "",
				privateKeyPassword: "",
			},
			wantAuth: func() *ssh.PublicKeys {
				privateKeyFile := ""
				u, err := user.Current()
				if err == nil {
					privateKeyFile = fmt.Sprintf("%s/.ssh/id_rsa", u.HomeDir)
				}
				gitAuth, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, "")
				return gitAuth
			}(),
			wantRepoUrl: "git@github.com:MR5356/syncer.git",
			wantErr:     false,
		},
		{
			name: "test token url",
			args: args{
				repo:               "https://token@github.com/MR5356/syncer.git",
				privateKeyFile:     "",
				privateKeyPassword: "",
			},
			wantAuth:    &http.TokenAuth{Token: "token"},
			wantRepoUrl: "https://github.com/MR5356/syncer.git",
			wantErr:     false,
		},
		{
			name: "test basic url",
			args: args{
				repo:               "https://username:password@github.com/MR5356/syncer.git",
				privateKeyFile:     "",
				privateKeyPassword: "",
			},
			wantAuth: &http.BasicAuth{
				Username: "username",
				Password: "password",
			},
			wantRepoUrl: "https://github.com/MR5356/syncer.git",
			wantErr:     false,
		},
		{
			name: "test access token url",
			args: args{
				repo:               "https://oauth2:access_token@github.com/MR5356/syncer.git",
				privateKeyFile:     "",
				privateKeyPassword: "",
			},
			wantAuth: &http.BasicAuth{
				Username: "oauth2",
				Password: "access_token",
			},
			wantRepoUrl: "https://github.com/MR5356/syncer.git",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAuth, gotRepoUrl, err := getAuth(tt.args.repo, tt.args.privateKeyFile, tt.args.privateKeyPassword)
			if (err != nil) != tt.wantErr {
				t.Errorf("getAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotAuth, tt.wantAuth) {
				t.Errorf("getAuth() gotAuth = %v, want %v", gotAuth, tt.wantAuth)
			}
			if gotRepoUrl != tt.wantRepoUrl {
				t.Errorf("getAuth() gotRepoUrl = %v, want %v", gotRepoUrl, tt.wantRepoUrl)
			}
		})
	}
}

func Test_getUrlType(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want urlType
	}{
		{
			name: "test git url",
			args: args{
				url: "git@github.com:MR5356/syncer.git",
			},
			want: gitUrlType,
		},
		{
			name: "test http url",
			args: args{
				url: "https://github.com/MR5356/syncer.git",
			},
			want: httpUrlType,
		},
		{
			name: "test token http url",
			args: args{
				url: "https://token@github.com/MR5356/syncer.git",
			},
			want: tokenizedHttpUrlType,
		},
		{
			name: "test basic http url",
			args: args{
				url: "https://username:password@github.com/MR5356/syncer.git",
			},
			want: basicHttpUrlType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUrlType(tt.args.url); got != tt.want {
				t.Errorf("getUrlType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sshKeyScan(t *testing.T) {
	type args struct {
		host string
		port int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test ssh key scan",
			args: args{
				host: "github.com",
				port: 22,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sshKeyScan(tt.args.host, tt.args.port)
		})
	}
}
