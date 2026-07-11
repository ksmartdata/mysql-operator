# chainsaw E2E（版本兼容性门禁）

目标：operator 每个 PR 在 5.7.44 / 8.0.37 /（8.4.9，8.4 适配落地后启用）三个版本上跑通
创建、my.cnf golden、复制、failover、更新配置五个场景，替代已废弃的 `test/e2e/`（ginkgo）。
版本矩阵依据：operator 代码路径只有 `Major == 5`、`8.0.x（< 8.4）`、`>= 8.4` 三个等价类，
每类取一个生产在用版本；8.0.31 与 8.0.37 走相同分支，不进 PR 门禁（可选 nightly）。

## 目录

```
test/e2e-chainsaw/
├── config/chainsaw-configuration.yaml   # chainsaw 全局配置（串行、超时、全局失败诊断）
├── values/mysql-<version>.yaml          # 版本参数（version/image/golden 路径），--values 注入
├── golden/                              # my.cnf golden 文件（步骤 0 在基线上生成，见下）
└── tests/
    ├── _shared/                         # 共享资源：CR/断言模板、create-cluster 步骤模板、sh 工具脚本
    ├── 01-create-replication/           # 场景 1+2+3：创建、my.cnf golden、复制
    ├── 02-failover/                     # 场景 4：强杀主库（崩溃 failover）→ 提升 → 旧主回归
    └── 03-config-update/                # 场景 5：patch mysqlConf → 滚动后回 Ready 且参数生效
```

三个 chainsaw Test 覆盖 06 文档的五个场景（1/2/3 合并共用一个集群，节省 CI 时长）。

## 本地运行

```bash
# 1. 构建 operator 镜像（必须从本地 checkout 构建：
#    arm64/images/mysql-operator/Dockerfile 会从 GitHub 重新下载
#    extra_image 分支源码，用它构建测不到本地改动）
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/mysql-operator_linux_amd64 ./cmd/mysql-operator
docker build -t mysql-operator:e2e -f hack/development/Dockerfile.operator .

# 2. kind 集群 + 部署 operator（与 CI 共用同一脚本，镜像可用
#    OPERATOR_IMAGE_REPO / OPERATOR_IMAGE_TAG 覆盖）
kind create cluster --name chainsaw
kind load docker-image mysql-operator:e2e --name chainsaw
./hack/e2e-chainsaw-setup.sh

# 3. 跑单版本
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
- 02 用例用 `--force --grace-period=0` 强杀主库：跳过 preStop 的
  `graceful-master-takeover-auto`，验证的才是 orchestrator 的 DeadMaster
  崩溃恢复路径；优雅删除走的是计划内主从切换，是另一条代码路径。

## 已知覆盖缺口（相对已废弃的 ginkgo 套件）

本套件当前只覆盖 06 文档的五个场景，旧 `test/e2e/`（ginkgo）曾覆盖而这里尚未覆盖的：

- 备份/恢复（MysqlBackup CR、init from backup）；
- 扩缩容（replicas 2→3→1）及缩容后 PVC 保留策略等不变量；
- unhealthy 节点从 service endpoints 摘除（只影响流量，不触发 failover 的场景）；
- `spec.readOnly: true` 只读集群；
- PVC 持久化路径（本套件全部 emptyDir）；
- GTID / read_only 变量级断言（旧套件直接断言 @@gtid_mode 等）。

按需补场景时优先复用 `tests/_shared/` 里的步骤模板与脚本。
