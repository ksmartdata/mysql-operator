# 修复步骤
- 确定所有pod都处于Running状态
- 进入任一pod的mysql容器执行:
   ```bash
   mysqlsh -uroot -proot_password
   dba.rebootClusterFromCompleteOutage()
   ```

   