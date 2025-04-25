package utils

import "testing"

func Test_getInstanceNameFromPodName(t *testing.T) {
	type args struct {
		podName string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "just one hyphen",
			args: args{
				podName: "mgr1-0",
			},
			want:    "mgr1",
			wantErr: false,
		},
		{
			name: "two hyphens",
			args: args{
				podName: "mgr-0304-0",
			},
			want:    "mgr-0304",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getInstanceNameFromPodName(tt.args.podName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getInstanceNameFromPodName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getInstanceNameFromPodName() got = %v, want %v", got, tt.want)
			}
		})
	}
}
