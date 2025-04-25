# 修复步骤
- 在ONLINE的SECONDARY的mysql容器内执行
  ```bash
  mysqlsh -uroot -p'root_password'
  \sql stop group_replication;
  \js dba.rebootClusterFromCompleteOutage()
  ```

   