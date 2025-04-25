package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	utilexec "k8s.io/utils/exec"
)

type ClusterStatus struct {
	ClusterName       string     `json:"clusterName"`
	DefaultReplicaSet ReplicaSet `json:"defaultReplicaSet"`
}

type ReplicaSet struct {
	Name       string               `json:"name"`
	Primary    string               `json:"primary"`
	Status     ReplicaSetStatus     `json:"status"`
	StatusText string               `json:"statusText"`
	Topology   map[string]*Instance `json:"topology"`
}

type ReplicaSetStatus string

const (
	ReplicaSetStatusOk            ReplicaSetStatus = "OK"
	ReplicaSetStatusOkPartial                      = "OK_PARTIAL"
	ReplicaSetStatusOkNoTolerance                  = "OK_NO_TOLERANCE"
	ReplicaSetStatusNoQuorum                       = "NO_QUORUM"
	ReplicaSetStatusUnknown                        = "UNKNOWN"
)

type Instance struct {
	MemberRole MemberRole     `json:"memberRole"`
	Address    string         `json:"address"`
	Mode       InstanceMode   `json:"mode"`
	Role       string         `json:"role"`
	Status     InstanceStatus `json:"status"`
}
type MemberRole string

const (
	PRIMARY   MemberRole = "PRIMARY"
	SECONDARY MemberRole = "SECONDARY"
)

type InstanceMode string

const (
	ReadWrite InstanceMode = "R/W"
	ReadOnly               = "R/O"
)

type InstanceStatus string

const (
	InstanceStatusOnline      InstanceStatus = "ONLINE"
	InstanceStatusMissing                    = "(MISSING)"
	InstanceStatusRecovering                 = "RECOVERING"
	InstanceStatusUnreachable                = "UNREACHABLE"
	InstanceStatusNotFound                   = ""
	InstanceStatusUnknown                    = "UNKNOWN"
)

const DefaultClusterName = "Cluster"

func sanitizeJSON(json []byte) []byte {
	return bytes.Replace(json, []byte("\\'"), []byte("'"), -1)
}

func GetClusterStatus(ctx context.Context, password, instanceName, ns string, index, timeout int) (*ClusterStatus, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()
	pythonStatement := "print(dba.get_cluster().status())"
	output, err := run(timeoutCtx, pythonStatement, password, instanceName, ns, index)
	if err != nil {
		return nil, err
	}
	if len(output.Stdout) == 0 {
		return nil, errors.New("get cluster status, no output")
	}
	status := &ClusterStatus{}
	err = json.Unmarshal(sanitizeJSON([]byte(output.Stdout)), status)
	if err != nil {
		return nil, errors.Wrapf(err, "decoding cluster status output: %q", output)
	}

	return status, nil
}

type ExecInfo struct {
	Command string
	Stdout  string
	Stderr  string
	Error   error
}

func run(ctx context.Context, pythonStatement, password, instanceName, ns string, index int) (*ExecInfo, error) {
	podName := GetPodNameFromInstanceName(instanceName, index)
	// mcamel-common-mgr-cluster-0.mcamel-common-mgr-cluster-instances.mcamel-system.svc.cluster.local:3306
	addr := fmt.Sprintf("%s.%s-instances.%s.svc.cluster.local", podName, instanceName, ns)
	stderr, stdout := &bytes.Buffer{}, &bytes.Buffer{}
	// mysqlsh --uri 'root:root123!@mcamel-common-kpanda-mgr-0.mcamel-common-kpanda-mgr-instances'
	uri := fmt.Sprintf("%s:%s@%s:%d", "root", escapePassword(password), addr, 3306)
	args := []string{"--no-wizard", "--uri", uri, "--py", "-e", pythonStatement}

	exec := utilexec.New()
	cmd := exec.CommandContext(ctx, "/usr/bin/mysqlsh", args...)

	cmd.SetStdout(stdout)
	cmd.SetStderr(stderr)
	execCmdPrefix := "exec command"
	fmt.Printf("%s: %s\n", execCmdPrefix, maskPassword(fmt.Sprintf("mysqlsh %s", strings.Join(args, " ")), escapePassword(password)))
	err := cmd.Run()
	res := &ExecInfo{
		Command: fmt.Sprintf("mysqlsh %s", strings.Join(args, " ")),
		Error:   err,
	}
	if err != nil {
		underlying := NewErrorFromStderr(stderr.String())
		if underlying != nil {
			return nil, errors.WithStack(underlying)
		}
	}
	res.Stdout = string(stripPasswordWarning(stdout.Bytes()))
	fmt.Printf("%sstdout: %s\n", getSpaces(execCmdPrefix, "stdout"), res.Stdout)
	if stderr != nil && stderr.String() != "" {
		res.Stderr = stderr.String()
		fmt.Printf("%sstderr: %s\n", getSpaces(execCmdPrefix, "stderr"), res.Stderr)
	}
	// 这里是为了把err在最后打印出来
	if err != nil {
		fmt.Printf("%serr: %s\n", getSpaces(execCmdPrefix, "err"), err)
		// 因为密码错误和没有PRIMARY的时候的信息其实是打印在stdout的, 所以这里附加到err里
		//err = appendError(err, res.Stdout, res.Stderr)
	}
	return res, err
}

// 所有的特殊字符: !"#$%&'()*+,-./:;<=>?@[\]^_`{|}~, 我们只允许: 数字、大小写字母、特殊符号!@%^*，至少 2 种；不能包含空格
func escapePassword(p string) string {
	return url.QueryEscape(p)
}

func appendError(err error, stdout, stderr string) error {
	if err == nil {
		return nil
	}
	otherInfo := ""
	if stderr != "" {
		otherInfo += stderr + ", "
	}
	if stdout != "" {
		otherInfo += stdout
	}
	if otherInfo == "" {
		return err
	}
	return fmt.Errorf("%s detail: %s", err, otherInfo)
}

func maskPassword(cmd, password string) string {
	mask := strings.Repeat("*", 6)
	return strings.Replace(cmd, password, mask, 1)
}
func getSpaces(s1, s2 string) string {
	if len(s1) < len(s2) {
		return ""
	}
	return strings.Repeat(" ", len(s1)-len(s2))
}
func GetPodNameFromInstanceName(instanceName string, index int) string {
	return fmt.Sprintf("%s-%d", instanceName, index)
}
func GetInstanceNameFromPodName(podName string) (string, error) {
	if podName == "" {
		return "", errors.New("env POD_NAME is empty?")
	}
	podNameParts := strings.Split(podName, "-")
	if len(podNameParts) < 2 {
		return "", fmt.Errorf("pod name should be xxx-0 format")
	}

	return strings.Join(podNameParts[:len(podNameParts)-1], "-"), nil
}

func stripPasswordWarning(in []byte) []byte {
	inStr := strings.Trim(string(in), "\n")
	keyWord := "insecure."
	index := strings.Index(inStr, keyWord)
	if index == -1 {
		return []byte(inStr)
	}
	length := len(keyWord)
	if len(inStr) < length {
		return []byte(inStr)
	}
	res := strings.Trim(inStr[index+length:], "\n")
	return []byte(res)
}

var errorRegex = regexp.MustCompile(`Traceback.*\n(?:  (.*)\n){1,}(?P<type>[\w\.]+)\: (?P<message>.*)`)

type Error struct {
	error
	Type    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func NewErrorFromStderr(stderr string) error {
	matches := errorRegex.FindAllStringSubmatch(stderr, -1)
	if len(matches) == 0 {
		return nil
	}
	result := make(map[string]string)
	for i, name := range errorRegex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = matches[len(matches)-1][i]
		}
	}
	return &Error{
		Type:    result["type"],
		Message: result["message"],
	}
}
