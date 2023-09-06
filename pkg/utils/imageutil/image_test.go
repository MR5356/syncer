package imageutil

import (
	"reflect"
	"testing"
)

func TestParseImageInfo(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    *ImageInfo
		wantErr bool
	}{
		{
			name: "test full name",
			args: args{
				src: "docker.ac.cn/library/nginx:latest",
			},
			want: &ImageInfo{
				Src:         "docker.ac.cn/library/nginx:latest",
				Registry:    "docker.ac.cn",
				Namespace:   "library",
				Project:     "nginx",
				TagOrDigest: "latest",
			},
			wantErr: false,
		},
		{
			name: "test short name",
			args: args{
				src: "nginx:latest",
			},
			want: &ImageInfo{
				Src:         "nginx:latest",
				Registry:    defaultRegistry,
				Namespace:   defaultNamespace,
				Project:     "nginx",
				TagOrDigest: "latest",
			},
			wantErr: false,
		},
		{
			name: "test digest name",
			args: args{
				src: "nginx@sha256:1fd62556954250bac80d601a196bb7fd480ceba7c10e94dd8fd4c6d1c08783d5",
			},
			want: &ImageInfo{
				Src:         "nginx@sha256:1fd62556954250bac80d601a196bb7fd480ceba7c10e94dd8fd4c6d1c08783d5",
				Registry:    defaultRegistry,
				Namespace:   defaultNamespace,
				Project:     "nginx",
				TagOrDigest: "sha256:1fd62556954250bac80d601a196bb7fd480ceba7c10e94dd8fd4c6d1c08783d5",
			},
			wantErr: false,
		},
		{
			name: "test full digest name",
			args: args{
				src: "docker.ac.cn/library/nginx@sha256:1fd62556954250bac80d601a196bb7fd480ceba7c10e94dd8fd4c6d1c08783d5",
			},
			want: &ImageInfo{
				Src:         "docker.ac.cn/library/nginx@sha256:1fd62556954250bac80d601a196bb7fd480ceba7c10e94dd8fd4c6d1c08783d5",
				Registry:    "docker.ac.cn",
				Namespace:   "library",
				Project:     "nginx",
				TagOrDigest: "sha256:1fd62556954250bac80d601a196bb7fd480ceba7c10e94dd8fd4c6d1c08783d5",
			},
			wantErr: false,
		},
		{
			name: "test full name without tag",
			args: args{
				src: "docker.ac.cn/library/nginx",
			},
			want: &ImageInfo{
				Src:         "docker.ac.cn/library/nginx",
				Registry:    "docker.ac.cn",
				Namespace:   "library",
				Project:     "nginx",
				TagOrDigest: "",
			},
			wantErr: false,
		},
		{
			name: "test full name without library",
			args: args{
				src: "docker.ac.cn/nginx:latest",
			},
			want: &ImageInfo{
				Src:         "docker.ac.cn/nginx:latest",
				Registry:    "docker.ac.cn",
				Namespace:   "",
				Project:     "nginx",
				TagOrDigest: "latest",
			},
			wantErr: false,
		},
		{
			name: "test full name without library and tag",
			args: args{
				src: "docker.ac.cn/nginx",
			},
			want: &ImageInfo{
				Src:         "docker.ac.cn/nginx",
				Registry:    "docker.ac.cn",
				Namespace:   "",
				Project:     "nginx",
				TagOrDigest: "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseImageInfo(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseImageInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseImageInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
