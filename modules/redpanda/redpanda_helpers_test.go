package redpanda

import (
	"testing"

	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

func Test_isAtLeastVersion(t *testing.T) {
	type args struct {
		image string
		major string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "v21.5.6",
			args: args{
				image: "redpandadata/redpanda:v21.5.6",
				major: "23.3",
			},
			want: false,
		},
		{
			name: "v23.3.3",
			args: args{
				image: "redpandadata/redpanda:v23.3.3",
				major: "23.3",
			},
			want: true,
		},
		{
			name: "v23.3.3-rc1",
			args: args{
				image: "redpandadata/redpanda:v23.3.3-rc1",
				major: "23.3",
			},
			want: true,
		},
		{
			name: "v21.3.3-rc1",
			args: args{
				image: "redpandadata/redpanda:v21.3.3-rc1",
				major: "23.3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Check(t, is.Equal(tt.want, isAtLeastVersion(tt.args.image, tt.args.major)), "isAtLeastVersion(%v, %v)", tt.args.image, tt.args.major)
		})
	}
}
