package types

import "testing"

func Test_parseTagOrDigest(t *testing.T) {
	type args struct {
		tagOrDigest string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test tag",
			args: args{
				tagOrDigest: "latest",
			},
			want: ":latest",
		},
		{
			name: "test digest",
			args: args{
				tagOrDigest: "sha256:1fd62556954250bac80d601a196bb7fd480ceba7c10e94dd8fd4c6d1c08783d5",
			},
			want: "@sha256:1fd62556954250bac80d601a196bb7fd480ceba7c10e94dd8fd4c6d1c08783d5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseTagOrDigest(tt.args.tagOrDigest); got != tt.want {
				t.Errorf("parseTagOrDigest() = %v, want %v", got, tt.want)
			}
		})
	}
}
