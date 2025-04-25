package main

import (
	"fmt"
	"strings"
)

const (
	IssueNetPartition                   = "检测到网络分区"
	IssueHasMultiplePrimaries           = "检测到多个主节点（脑裂）"
	IssueNoPrimary                      = "集群无主节点"
	IssueNoPrimaryButHasSecondaryOnline = "集群无主节点,有SECONDARY节点在"
	IssueSecondaryQuorumUnmet           = "集群SECONDARY节点在线数量不足"
)

var (
	issueSolutions = map[IssueType][]string{
		IssueNetPartition:                   {networkPartition, checkConnect},
		IssueHasMultiplePrimaries:           {multiplyPrimary, checkConnect},
		IssueNoPrimary:                      {noPrimary},
		IssueNoPrimaryButHasSecondaryOnline: {noPrimaryButHasSecondaryOnline},
		IssueSecondaryQuorumUnmet:           {secondaryQuorumUnmet},
	}
)

type IssueType string

func (i IssueType) String() string {
	return string(i)
}

func (i IssueType) Wrapped() IssueType {
	return IssueType(fmt.Sprintf("⚠️%s", i))
}

func (i IssueType) GetSolutions() string {
	solutions := issueSolutions[i]
	return fmt.Sprintf(strings.Join(solutions, "\n"))
}

func (i IssueType) Print() {
	fmt.Println(i.Wrapped())
	fmt.Println(i.GetSolutions())
}

func ConvertFromString(s string) IssueType {
	return IssueType(s)
}

func HandleError(s string) string {
	if strings.TrimSpace(s) == "context deadline exceeded" {
		return fmt.Sprintf("连接超时: 设置更大的--timeout参数或者: \n%s", checkConnect)
	}
	return fmt.Sprintf("错误: %s, 请按照错误提示, 处理相应的错误", s)
}
