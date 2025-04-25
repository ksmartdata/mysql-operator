**检查2个节点之间的连通性**
在mysql这个container里执行
- 如果能连通, 会出现msyql的界面
  ```bash
  // 在mcamel-common-mgr-cluster-0 节点上测试与 mcamel-common-mgr-cluster-2的连通性
  bash-5.1$ mysql -uroot -p'root_password' -hmcamel-common-mgr-cluster-2.mcamel-common-mgr-cluster-instances.mcamel-system.svc.cluster.local
  ```
  
- 如果不能, 应该有错误信息
  ```bash
  bash-5.1$ mysql -uroot -p'root_password' -hmcamel-common-mgr-cluster-2.mcamel-common-mgr-cluster-instances.mcamel-system.svc.cluster.local
  mysql: [Warning] Using a password on the command line interface can be insecure.
  ERROR 2005 (HY000): Unknown MySQL server host 'mcamel-common-mgr-cluster-2.mcamel-common-mgr-cluster-instances.mcamel-system.svc.cluster.local' (-2)
  ```
  