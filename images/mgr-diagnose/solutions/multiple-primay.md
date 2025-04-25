# 修复步骤
- **确定所有pod都处于Running状态**
- **确认脑裂情况**
   一般这种情况在2个节点执行如下命令,这2个值应该是不一样的
   ```javascript
   // 分别连接到不同的主节点检查状态
   bash-5.1$ mysqlsh -uroot -p'root_password' --py -e 'print(dba.get_cluster().name)'
   name1
   
   bash-5.1$ mysqlsh -uroot -p'root_password' --py -e 'print(dba.get_cluster().name)'
   name2
   ```
- **停止写入操作**
    * 立即停止所有应用程序对数据库的写入操作，防止数据进一步分歧: 将对应的router的deployment的replicas设置为0
   ```bash
   > kubectl get deploy -n mcamel-system|grep router
   mcamel-common-kpanda-mgr-router            1/1     1            1           8d
   mcamel-common-mgr-cluster-router           1/1     1            1           8d

   > kubectl scale deploy mcamel-common-mgr-cluster-router --replicas=0
   deployment.apps/mcamel-common-mgr-cluster-router scaled
   ```
- **确定这2个primary是能够相互连通的**
  
  参看: **检查2个节点之间的连通性**
- **确定保留哪个主节点**
   ```bash
   // 检查各节点的事务计数

   // primary1
   bash-5.1$ mysqlsh -uroot -proot_password6 
   \sql SELECT @@GLOBAL.GTID_EXECUTED;

   // primary2
   bash-5.1$ mysqlsh -uroot -proot_password
   \sql SELECT @@GLOBAL.GTID_EXECUTED;
   ```
  假设有两个主节点，执行 SELECT @@GLOBAL.GTID_EXECUTED; 后得到以下结果：
  
  主节点 1 (primary1):
  ```text
  3a01df00-234a-11e9-a7d3-080027c2be11:1-100,
  5bc89f6f-234a-11e9-a7d3-080027c2be11:1-50
  ```
  主节点 2 (primary2):
  ```text
  3a01df00-234a-11e9-a7d3-080027c2be11:1-95,
  5bc89f6f-234a-11e9-a7d3-080027c2be11:1-50,
  7de31f45-234a-11e9-a7d3-080027c2be11:1-10
  ```
  检查 primary2 的 GTID 是否完全被 primary1 包含
  返回 1 表示 primary2_gtid ⊆ primary1_gtid（primary1_gtid包含 primary2_gtid ），否则返回 0。
  ```text
  # 如果互相GTID_SUBSET都不返回1,就说明GTID有冲突,需要业务上确认.
  SELECT GTID_SUBSET('primary2_gtid', 'primary1_gtid')\G
   
  SELECT GTID_SUBSET('primary1_gtid', 'primary2_gtid')\G
  ```
- **停止不保留的主节点的组复制**
   ```bash
   // 连接到不保留的主节点priamry2
   bash-5.1$ mysqlsh -uroot -p'root_password' --py -e 'dba.get_cluster().dissolve({"force": True})'
   Are you sure you want to dissolve the cluster? [y/N]: y
   ```
- **重建集群**
   ```javascript
   // 连接到保留的主节点primary1
   mysqlsh -uroot -proot_password

   cluster = dba.getCluster()

   // 将之前的主节点作为从节点重新加入
   cluster.addInstance("primary2:3306", {recoveryMethod: "clone"})
   ```
