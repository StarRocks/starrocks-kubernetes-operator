package common

import "testing"

func TestEqualsIgnoreCase(t *testing.T) {
	type args struct {
		a string
		b string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test equals ignore case",
			args: args{
				a: "test",
				b: "TeSt",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EqualsIgnoreCase(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("EqualsIgnoreCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
