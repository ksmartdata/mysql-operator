# DEPRECATED

本目录下的 ginkgo e2e 套件（bitpoke 上游遗留）已废弃，不再维护、不做新版本适配：

- 年久失修，且用例中包含 MySQL 8.4 已移除的旧复制语法，无法覆盖新版本；
- 实际不在任何 CI 上运行（仓库根目录的 `.drone.yml` 里有一条引用它的
  "e2e testing" pipeline，是 bitpoke 上游遗留，依赖其 GCP secrets 与
  drone 环境，在本 fork 上不生效）。

替代方案：`test/e2e-chainsaw/`（kyverno chainsaw 声明式 E2E），由
`.github/workflows/e2e-chainsaw.yml` 在每个 PR 上按版本矩阵运行，
详见 [test/e2e-chainsaw/README.md](../e2e-chainsaw/README.md)。
