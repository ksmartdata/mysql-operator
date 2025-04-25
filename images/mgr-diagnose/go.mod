module daocloud.io/mcamel/mgr-diagnose

go 1.23.3

replace github.com/oracle/mysql-operator => /root/daocloud/mcamel/oracle-golang-mysql-operator

require (
	github.com/pkg/errors v0.9.1
	k8s.io/utils v0.0.0-20241210054802-24370beab758
)
