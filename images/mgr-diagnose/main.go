package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"reflect"

	"daocloud.io/mcamel/mgr-diagnose/utils"
)

var (
	//go:embed solutions/multiple-primay.md
	multiplyPrimary string
	//go:embed solutions/check-connect.md
	checkConnect string
	//go:embed solutions/no-primary.md
	noPrimary string
	//go:embed solutions/no-primary-but-has-secondary-online.md
	noPrimaryButHasSecondaryOnline string
	//go:embed solutions/secondary-quorum-unmet.md
	secondaryQuorumUnmet string
	//go:embed solutions/network-partition.md
	networkPartition string
)

type PodStatus struct {
	PodName string
	Status  *utils.ClusterStatus
	Error   error
}

func main() {
	var (
		password         string
		expectedReplicas int
		instanceName     string
		timeout          int
	)

	flag.StringVar(&password, "password", "", "MySQL root密码")
	flag.IntVar(&expectedReplicas, "replicas", 0, "期望副本数")
	flag.IntVar(&timeout, "timeout", 10, "访问节点超时时间（秒）")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "MySQL MGR 集群诊断工具\n\n用法：%s [选项]\n\n选项：\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\n示例：\n  %s --password=123456 --replicas=3 --timeout=20\n", os.Args[0])
	}

	flag.Parse()

	// 当没有参数时主动显示帮助
	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}
	namespace := getEnv("POD_NAMESPACE")
	instanceName, err := utils.GetInstanceNameFromPodName(getEnv("POD_NAME"))
	if err != nil {
		panic(err)
	}
	// 验证必需参数
	if password == "" || expectedReplicas == 0 {
		fmt.Println("错误：缺少必需参数")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if timeout <= 0 {
		fmt.Println("错误：timeout应该为正整数")
		os.Exit(1)
	}

	// 获取所有Pod列表
	results := make(chan PodStatus, expectedReplicas)
	for i := 0; i < expectedReplicas; i++ {
		status, err := utils.GetClusterStatus(context.Background(), password, instanceName, namespace, i, timeout)
		results <- PodStatus{utils.GetPodNameFromInstanceName(instanceName, i), status, err}
		fmt.Printf("##############################################################\n\n")
	}
	close(results)

	// 分析结果
	analyzeResults(results, expectedReplicas)
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("env %s is empty, check the result of 'echo $%s' in sidecar's shell", key, key))
	}
	return value
}

// 生成Pod列表（根据实际命名规则调整）
func generatePodList(instance, namespace string, replicas int) []string {
	var pods []string
	for i := 0; i < replicas; i++ {
		pods = append(pods, fmt.Sprintf("%s-%d", instance, i))
	}
	return pods
}

// 分析聚合结果
func analyzeResults(results <-chan PodStatus, expectedReplicas int) {
	var (
		allIssues             []IssueType
		primaryOnlineRecord   = make(map[string]int)
		secondaryOnlineRecord = make(map[string]int)
	)
	prevPodStatus := PodStatus{}
	// 收集所有结果
	for item := range results {
		if prevPodStatus.PodName != "" {
			if prevPodStatus.Error == nil && item.Error == nil {
				if !reflect.DeepEqual(prevPodStatus.Status, item.Status) {
					fmt.Printf("❌ 可能出现网络分区, Pod %s 与 %s 的输出不一致\n\n", item.PodName, prevPodStatus.PodName)
					fmt.Println(networkPartition, "\n", checkConnect)
					return
				}
			}
		} else {
			prevPodStatus = item
		}
		if item.Error != nil {
			info := fmt.Sprintf("Pod %s 检测失败: %v", item.PodName, item.Error)
			fmt.Println(HandleError(info))
			allIssues = append(allIssues, ConvertFromString(info))
			continue
		}

		status := item.Status
		// 统计主节点
		if status.DefaultReplicaSet.Primary != "" {
			// 多个PRIMARY, 如果一个在线, 一个不在线, 在前面就被判定为网络分区了
			primaryNode := status.DefaultReplicaSet.Topology[status.DefaultReplicaSet.Primary]
			if primaryNode.MemberRole == utils.PRIMARY && primaryNode.Status == utils.InstanceStatusOnline {
				primaryOnlineRecord[status.DefaultReplicaSet.Primary]++
			}
		}

		// 统计在线节点
		for _, node := range status.DefaultReplicaSet.Topology {
			if node.Status == utils.InstanceStatusOnline && node.MemberRole == utils.SECONDARY {
				secondaryOnlineRecord[node.Address]++
			}
		}
	}
	primaryOnlineCount := len(primaryOnlineRecord)
	secondaryOnlineCount := len(secondaryOnlineRecord)
	var issue IssueType
	// 1. 主节点一致性检查
	if primaryOnlineCount > 1 {
		issue = IssueHasMultiplePrimaries
		issue.Print()
		return
	} else if primaryOnlineCount == 0 {
		if secondaryOnlineCount != 0 {
			issue = IssueNoPrimaryButHasSecondaryOnline
		} else {
			issue = IssueNoPrimary
		}
		issue.Print()
		return
	}

	// 2. 在线节点检查
	// 应该只有SECONDARY节点不在线
	if secondaryOnlineCount < expectedReplicas-1 {
		issue = IssueSecondaryQuorumUnmet
		issue.Print()
		return
	}

	if len(allIssues) == 0 {
		fmt.Println("✅ 所有集群状态正常")
		return
	}
}
