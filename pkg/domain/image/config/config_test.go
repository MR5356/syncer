package config

import (
	"github.com/MR5356/syncer/pkg/utils/structutil"
	"reflect"
	"testing"
)

func TestNewConfigFromFile(t *testing.T) {
	type args struct {
		cf string
	}
	tests := []struct {
		name string
		args args
		want *Config
	}{
		{
			name: "NewConfigFromYaml",
			args: args{
				cf: "testdata/config.yaml",
			},
			want: &Config{
				Auth: map[string]*Auth{
					"docker.io": {
						Username: "user",
						Password: "passwd",
						Insecure: false,
					},
				},
				Images: map[string]any{
					"nginx:latest": "test/nginx:latest",
				},
				Proc:    8,
				Retries: 10,
			},
		},
		{
			name: "NewConfigFromJson",
			args: args{
				cf: "testdata/config.json",
			},
			want: &Config{
				Auth: map[string]*Auth{
					"docker.io": {
						Username: "user",
						Password: "passwd",
						Insecure: false,
					},
				},
				Images: map[string]any{
					"nginx:latest": "test/nginx:latest",
				},
				Proc:    8,
				Retries: 10,
			},
		},
		{
			name: "NewConfigFromYaml With empty",
			args: args{
				cf: "testdata/config_nothing.yaml",
			},
			want: &Config{
				Auth:    make(map[string]*Auth),
				Images:  make(map[string]any),
				Proc:    0,
				Retries: 0,
			},
		},
		{
			name: "NewConfigFromJson With empty",
			args: args{
				cf: "testdata/config_nothing.json",
			},
			want: &Config{
				Auth:    make(map[string]*Auth),
				Images:  make(map[string]any),
				Proc:    0,
				Retries: 0,
			},
		},
		{
			name: "NewConfigFromJson With some empty",
			args: args{
				cf: "testdata/config_some.json",
			},
			want: &Config{
				Auth: make(map[string]*Auth),
				Images: map[string]any{
					"nginx:latest": "test/nginx:latest",
				},
				Proc:    8,
				Retries: 10,
			},
		},
		{
			name: "NewConfigFromYaml With some empty",
			args: args{
				cf: "testdata/config_some.yaml",
			},
			want: &Config{
				Auth: make(map[string]*Auth),
				Images: map[string]any{
					"nginx:latest": "test/nginx:latest",
				},
				Proc:    8,
				Retries: 10,
			},
		},
		{
			name: "NewConfigFromYaml With image dest array",
			args: args{
				cf: "testdata/config_images_dest_array.yaml",
			},
			want: &Config{
				Auth: make(map[string]*Auth),
				Images: map[string]any{
					"nginx:latest": []string{
						"test/nginx:latest",
						"test2/nginx:latest",
					},
				},
				Proc:    8,
				Retries: 10,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConfigFromFile(tt.args.cf); !reflect.DeepEqual(structutil.Struct2String(got), structutil.Struct2String(tt.want)) {
				t.Errorf("NewConfigFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
