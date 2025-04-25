### tip
 ```shell
 ksmartdata/mysql-operator-orchestrator   v0.6.2 
 ksmartdata/mysql-operator                v0.6.2 
 ksmartdata/mysql-operator-sidecar-8.0    v0.6.2  
 ksmartdata/mysql-operator-sidecar-5.7    v0.6.2
 ksmartdata/logical-backup                  
 ```

### logical-backup

* 拷贝 postgres-operator 仓库中的原始文件，并增加在上面包一层 delete，用于支持删除操作。
* 新增环境变量 LOGICAL_BACKUP_FILE_NAME，用于指定备份文件名，如果不指定，则使用默认的备份文件名(date +%s).