package v1

import "testing"

func TestAutoScalerVersion_Complete(t *testing.T) {
	type args struct {
		major string
		minor string
	}
	tests := []struct {
		name    string
		version AutoScalerVersion
		args    args
		want    AutoScalerVersion
	}{
		{
			name:    "test for version v1",
			version: AutoScalerV1,
			args:    args{},
			want:    AutoScalerV1,
		},
		{
			name:    "test for version v2",
			version: AutoScalerV2,
			args:    args{},
			want:    AutoScalerV2,
		},
		{
			name:    "test for version v2beta2",
			version: AutoScalerV2Beta2,
			args:    args{},
			want:    AutoScalerV2Beta2,
		},
		{
			name:    "test for empty version",
			version: "",
			args: args{
				major: "1",
				minor: "27",
			},
			want: AutoScalerV2,
		},
		{
			name:    "test for empty version2",
			version: "",
			args: args{
				major: "1",
				minor: "25",
			},
			want: AutoScalerV2Beta2,
		},
		{
			name:    "test for empty version3",
			version: "",
			args: args{
				major: "1",
				minor: "23",
			},
			want: AutoScalerV2Beta2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.Complete(tt.args.major, tt.args.minor); got != tt.want {
				t.Errorf("Complete() = %v, want %v", got, tt.want)
			}
		})
	}
}
