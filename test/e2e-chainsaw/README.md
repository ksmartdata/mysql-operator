# chainsaw E2E（版本兼容性门禁）

目标：operator 每个 PR 在 5.7.44 / 8.0.37 /（8.4.9，8.4 适配落地后启用）三个版本上跑通
创建、my.cnf golden、复制、failover、更新配置五个场景，替代已废弃的 `test/e2e/`（ginkgo）。
版本矩阵依据：operator 代码路径只有 `Major == 5`、`8.0.x（< 8.4）`、`>= 8.4` 三个等价类，
每类取一个生产在用版本；8.0.31 与 8.0.37 走相同分支，不进 PR 门禁（可选 nightly）。

## 目录

```
test/e2e-chainsaw/
├── config/chainsaw-configuration.yaml   # chainsaw 全局配置（串行、超时）
├── values/mysql-<version>.yaml          # 版本参数（version/image/golden 路径），--values 注入
├── golden/                              # my.cnf golden 文件（步骤 0 在基线上生成，见下）
└── tests/
    ├── 01-create-replication/           # 场景 1+2+3：创建、my.cnf golden、复制
    ├── 02-failover/                     # 场景 4：kill 主库 → 提升 → 旧主回归
    └── 03-config-update/                # 场景 5：patch mysqlConf → 滚动后回 Ready
```

三个 chainsaw Test 覆盖 06 文档的五个场景（1/2/3 合并共用一个集群，节省 CI 时长）。

## 本地运行

```bash
# 1. kind 集群 + 部署 operator（operator 镜像按需替换）
kind create cluster --name chainsaw
docker build -t mysql-operator:e2e -f arm64/images/mysql-operator/Dockerfile .
kind load docker-image mysql-operator:e2e --name chainsaw
# podSecurityContext=null 对齐 mcamel 生产 chart（默认 runAsUser 65532 会让
# orchestrator 容器写 /etc/orchestrator 配置被拒而 CrashLoop）
helm install mysql-operator ./deploy/charts/mysql-operator \
  --set podSecurityContext=null \
  --set image.repository=mysql-operator --set image.tag=e2e \
  --set orchestrator.image.repository=ghcr.io/ksmartdata/mysql-operator-orchestrator \
  --set orchestrator.image.tag=v0.7.3 \
  --set sidecar57.image.repository=ghcr.io/ksmartdata/mysql-operator-sidecar-5.7 \
  --set sidecar57.image.tag=v0.7.4-1 \
  --set sidecar80.image.repository=ghcr.io/ksmartdata/mysql-operator-sidecar-8.0 \
  --set sidecar80.image.tag=v0.7.5-1

# 2. 跑单版本
chainsaw test --test-dir test/e2e-chainsaw/tests \
  --config test/e2e-chainsaw/config/chainsaw-configuration.yaml \
  --values test/e2e-chainsaw/values/mysql-5.7.44.yaml
```

## 步骤 0：生成 golden（改代码前必须先做）

golden 是"旧路径 diff 为零"的机械化执行：5.7.44 / 8.0.37 的 golden 必须取自
**改造前基线**。本分支不含任何 operator 代码改动，CI 构建产物即基线，因此步骤 0
可直接通过 CI 完成：golden 缺失时 01 用例会把实际 my.cnf 全文打印在日志的
`-----BEGIN ACTUAL MY.CNF-----` 标记之间，从 CI 日志采集入库即可：

```bash
# 或本地生成（基线环境创建 tests/_shared/cluster.yaml 同款集群后）：
kubectl get cm e2e-mysql -n <ns> -o jsonpath='{.data.my\.cnf}' \
  > test/e2e-chainsaw/golden/my.cnf-5.7.44.cnf   # 8.0.37 同理
```

golden 入库后 01 用例的 golden 步骤才会通过。**8.4.9 的 golden 在 8.4 功能落地时生成并随代码评审。**

## 注意

- **测试 SQL 用 `sys_operator` 账号**（密码从 `e2e-mysql-operated` secret 现取），不用 root；
  测试数据写 `sys_operator` 库（该账号在 `*.*` 上无 INSERT/DELETE，对工具库有 ALL）。
  原因：`mysql:5.7` 镜像在磁盘性能差的机器上首启初始化会失败——entrypoint 探活客户端
  连不上实际已就绪的临时 server（CI 实测 run 29088678923：30 次全败 → `Unable to start
  server` → 容器重启后跳过初始化），root 密码永远不会被设置（空密码反而可登录）。这是
  上游 docker-library/mysql 的已知类型问题；5.7.44 是 5.7 最后一个镜像（2023-10 停更），
  不会再修。8.0+ 的新版 entrypoint 未复现。sys_operator 由 init_file 每次启动重建，
  不受影响。生产节点磁盘较好通常不中招，但值得抽查存量 5.7.44 实例的 root 可用性。

- CR 模板与 mcamel 下发同形：同时显式 `spec.mysqlVersion`（完整版本号）+ `spec.image`
  （mcamel 生产镜像为 `library/mysql:{Version}` 社区镜像，CI 直连 docker.io）。
- 用例存储用 emptyDir（只测 operator 行为，不测持久化）；mcamel 生产为 PVC。
- 断言里的 JMESPath 表达式（conditions 按 type 过滤）在首次真实运行时可能需微调，
  以 chainsaw 实际报错为准。
