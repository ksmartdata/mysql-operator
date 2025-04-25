package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"daocloud.io/mcamel/mgr-diagnose/utils"
	"github.com/pkg/errors"
)

func getTemplate() utils.ClusterStatus {
	// 读取test-template.json
	template, err := os.ReadFile("test-template.json")
	if err != nil {
		panic(err)
	}
	res := new(utils.ClusterStatus)
	if err = json.Unmarshal(template, res); err != nil {
		panic(err)
	}
	return *res
}

func Test_analyzeResults(t *testing.T) {
	type args struct {
		results          <-chan PodStatus
		expectedReplicas int
		f                func(replicas int) <-chan PodStatus
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "cluster-is-ok",
			args: args{
				expectedReplicas: 3,
				f: func(replicas int) <-chan PodStatus {
					res := make(chan PodStatus, replicas)
					for i := 0; i < replicas; i++ {
						t := getTemplate()
						res <- PodStatus{
							PodName: fmt.Sprintf("mgr0304-%d", i),
							Status:  &t,
							Error:   nil,
						}
					}
					close(res)
					return res
				},
			},
		},
		{
			// 就是两个pod的输出不一致, 包括多个PRIMARY: 有的pod的PRIMARY在线, 有的pod的PRIMARY不在线
			name: "network-partition",
			args: args{
				expectedReplicas: 3,
				f: func(replicas int) <-chan PodStatus {
					res := make(chan PodStatus, replicas)
					for i := 0; i < replicas; i++ {
						t := getTemplate()
						podName := fmt.Sprintf("mgr0304-%d", i)
						if i == 1 {
							t.DefaultReplicaSet.Status = utils.ReplicaSetStatusOkPartial
						}
						res <- PodStatus{
							PodName: podName,
							Status:  &t,
							Error:   nil,
						}
					}
					close(res)
					return res
				},
			},
		},
		{
			// 这种情况也不容易出现
			name: "no primary, has secondaries",
			args: args{
				expectedReplicas: 3,
				f: func(replicas int) <-chan PodStatus {
					res := make(chan PodStatus, replicas)
					for i := 0; i < replicas; i++ {
						t := getTemplate()
						t.DefaultReplicaSet.Topology["mgr0304-0.mgr0304-instances.default.svc.cluster.local:3306"].Status = utils.InstanceStatusMissing
						res <- PodStatus{
							PodName: fmt.Sprintf("mgr0304-%d", i),
							Status:  &t,
							Error:   nil,
						}
					}
					close(res)
					return res
				},
			},
		},
		{
			// 正常的应该是这种
			name: "no primary, no secondary",
			args: args{
				expectedReplicas: 3,
				f: func(replicas int) <-chan PodStatus {
					res := make(chan PodStatus, replicas)
					for i := 0; i < replicas; i++ {
						res <- PodStatus{
							PodName: fmt.Sprintf("mgr0304-%d", i),
							Status:  nil,
							Error:   fmt.Errorf("exit status 1"),
						}
					}
					close(res)
					return res
				},
			},
		},
		{
			// 会因为对比两个pod的输出就之间判断为网络分区了
			name: "multiple-primaries",
			args: args{
				expectedReplicas: 3,
				f: func(replicas int) <-chan PodStatus {
					t := getTemplate()
					res := make(chan PodStatus, replicas)
					for i := 0; i < replicas; i++ {
						podName := fmt.Sprintf("mgr0304-%d", i)
						if i == 0 {
							t1 := getTemplate()
							t1.DefaultReplicaSet.Primary = "mgr0304-1.mgr0304-instances.default.svc.cluster.local:3306"
							t1.DefaultReplicaSet.Topology["mgr0304-1.mgr0304-instances.default.svc.cluster.local:3306"].MemberRole = utils.PRIMARY
							res <- PodStatus{
								PodName: podName,
								Status:  &t1,
							}
						} else {
							res <- PodStatus{
								PodName: podName,
								Status:  &t,
							}
						}

					}
					close(res)
					return res
				},
			},
		},
		{
			name: "multiple-primaries-has-error",
			args: args{
				expectedReplicas: 3,
				f: func(replicas int) <-chan PodStatus {
					res := make(chan PodStatus, replicas)
					for i := 0; i < replicas; i++ {
						podName := fmt.Sprintf("mgr0304-%d", i)
						if i == 0 {
							t1 := getTemplate()
							t1.DefaultReplicaSet.Primary = "mgr0304-1.mgr0304-instances.default.svc.cluster.local:3306"
							t1.DefaultReplicaSet.Topology["mgr0304-1.mgr0304-instances.default.svc.cluster.local:3306"].MemberRole = utils.PRIMARY
							res <- PodStatus{
								PodName: podName,
								Status:  &t1,
							}
						} else {
							res <- PodStatus{
								PodName: podName,
								Status:  nil,
								Error:   errors.New("some error"),
							}
						}

					}
					close(res)
					return res
				},
			},
		},
		{
			name: "no-primary-has-errors",
			args: args{
				expectedReplicas: 3,
				f: func(replicas int) <-chan PodStatus {
					res := make(chan PodStatus, replicas)
					for i := 0; i < replicas; i++ {
						res <- PodStatus{
							PodName: fmt.Sprintf("mgr0304-%d", i),
							Status:  nil,
							Error:   fmt.Errorf("some error"),
						}
					}
					close(res)
					return res
				},
			},
		},
		{
			name: "has-primary-has-errors",
			args: args{
				expectedReplicas: 3,
				f: func(replicas int) <-chan PodStatus {
					res := make(chan PodStatus, replicas)
					for i := 0; i < replicas; i++ {
						t := getTemplate()
						if i == 0 {
							res <- PodStatus{
								PodName: fmt.Sprintf("mgr0304-%d", i),
								Status:  nil,
								Error:   fmt.Errorf("some error"),
							}
						} else {
							res <- PodStatus{
								PodName: fmt.Sprintf("mgr0304-%d", i),
								Status:  &t,
								Error:   nil,
							}
						}

					}
					close(res)
					return res
				},
			},
		},
		{
			name: "secondary unmet",
			args: args{
				expectedReplicas: 3,
				f: func(replicas int) <-chan PodStatus {
					res := make(chan PodStatus, replicas)
					for i := 0; i < replicas; i++ {
						t := getTemplate()
						t.DefaultReplicaSet.Topology["mgr0304-1.mgr0304-instances.default.svc.cluster.local:3306"].Status = utils.InstanceStatusMissing
						res <- PodStatus{
							PodName: fmt.Sprintf("mgr0304-%d", i),
							Status:  &t,
							Error:   nil,
						}
					}
					close(res)
					return res
				},
			},
		},
		{
			name: "secondary unmet, has errors",
			args: args{
				expectedReplicas: 3,
				f: func(replicas int) <-chan PodStatus {
					res := make(chan PodStatus, replicas)
					for i := 0; i < replicas; i++ {
						t := getTemplate()
						t.DefaultReplicaSet.Topology["mgr0304-1.mgr0304-instances.default.svc.cluster.local:3306"].Status = utils.InstanceStatusMissing
						if i == 0 {
							res <- PodStatus{
								PodName: fmt.Sprintf("mgr0304-%d", i),
								Status:  nil,
								Error:   fmt.Errorf("some error"),
							}
						} else {
							res <- PodStatus{
								PodName: fmt.Sprintf("mgr0304-%d", i),
								Status:  &t,
								Error:   nil,
							}
						}
					}
					close(res)
					return res
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.results = tt.args.f(tt.args.expectedReplicas)
			analyzeResults(tt.args.results, tt.args.expectedReplicas)
		})
	}
}
