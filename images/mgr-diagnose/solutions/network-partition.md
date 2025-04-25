**处理网络分区**
- 确定所有pod都处于Running状态
- 查看任2个pod之间网络连接情况(参见: 检查2个节点之间的连通性), 查看coredns pod的日志, 查看MGR实例的svc和ep资源是否指向了正确的pod,是否有networkpolicy之类的网络设置.
