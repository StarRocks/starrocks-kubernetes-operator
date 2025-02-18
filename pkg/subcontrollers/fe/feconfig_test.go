package fe

import "testing"

func TestIsRunInSharedDataMode(t *testing.T) {
	type args struct {
		config map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "shared data mode",
			args: args{
				config: map[string]interface{}{
					"run_mode": "shared_data",
				},
			},
			want: true,
		},
		{
			name: "shared nothing mode - 1",
			args: args{
				config: map[string]interface{}{},
			},
			want: false,
		},
		{
			name: "shared nothing mode - 2",
			args: args{
				config: map[string]interface{}{
					"run_mode": "shared_nothing",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRunInSharedDataMode(tt.args.config); got != tt.want {
				t.Errorf("IsRunInSharedDataMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
