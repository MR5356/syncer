package structutil

import "testing"

func TestStruct2String(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test int",
			args: args{v: 1},
			want: "1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Struct2String(tt.args.v); got != tt.want {
				t.Errorf("Struct2String() = %v, want %v", got, tt.want)
			}
		})
	}
}
