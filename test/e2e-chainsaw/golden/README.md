# my.cnf golden 文件

此目录存放各版本 my.cnf ConfigMap 的期望内容（逐字节比对），是
"存量版本配置 diff 为零"规则的机械化执行——operator 任何改动若无意间
改变了 5.7/8.0 集群生成的 my.cnf，01 用例的 golden 步骤直接红灯。

生成方式见上层 README「步骤 0」。**golden 必须取自改造前基线**，
不要在功能分支上重新生成来"让测试变绿"——那等于把回归写进期望值。
更新 golden 的唯一合法场景：有意变更配置且在 MR 里说明并评审。

**golden 与 CR 资源规格耦合**：operator 会按 pod memory 推导 `innodb-buffer-pool-size` /
`innodb-log-file-size` 写进 my.cnf（追加在文件尾部）。改动 `tests/_shared/cluster.yaml`
的 resources 时必须同步更新 golden（CI 实测：512Mi→768Mi 新增了 buffer-pool 行）。

- my.cnf-5.7.44.cnf —— 已入库（基线 run 29083492185 + 768Mi 资源修正）
- my.cnf-8.0.37.cnf —— 已入库（同上）
- my.cnf-8.4.9.cnf  —— 8.4 适配功能落地时生成并评审
