- **尝试rejoin节点**
  ```bash
  // 连接到主节点
  mysqlsh -uroot -proot_password
    
  cluster = dba.getCluster()
    
  // 尝试重新加入节点
  cluster.rejoinInstance("secondary:3306")
  ```

- **如果数据差异较大或者rejoinInstance失败**
  ```bash
  // 连接到主节点
  mysqlsh -uroot -proot_password
    
  cluster = dba.getCluster()
    
  // 使用克隆方法重建节点
  cluster.removeInstance("secondary:3306", {force: true})
  cluster.addInstance("secondary:3306", {recoveryMethod: "clone"})
  ```

