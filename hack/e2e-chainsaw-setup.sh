#!/usr/bin/env bash
# 部署 e2e 用 mysql-operator chart（CI 与本地共用同一份命令，避免两处漂移）
set -euo pipefail

OPERATOR_IMAGE_REPO="${OPERATOR_IMAGE_REPO:-mysql-operator}"
OPERATOR_IMAGE_TAG="${OPERATOR_IMAGE_TAG:-e2e}"

# podSecurityContext=null 对齐 mcamel 生产 chart：默认的 runAsUser 65532
# 会让 orchestrator 容器写 /etc/orchestrator/orchestrator.conf.json 被拒（CrashLoop）
helm install mysql-operator ./deploy/charts/mysql-operator \
  --set podSecurityContext=null \
  --set image.repository="$OPERATOR_IMAGE_REPO" \
  --set image.tag="$OPERATOR_IMAGE_TAG" \
  --set image.pullPolicy=IfNotPresent \
  --set orchestrator.image.repository=ghcr.io/ksmartdata/mysql-operator-orchestrator \
  --set orchestrator.image.tag=v0.7.3 \
  --set sidecar57.image.repository=ghcr.io/ksmartdata/mysql-operator-sidecar-5.7 \
  --set sidecar57.image.tag=v0.7.4-1 \
  --set sidecar80.image.repository=ghcr.io/ksmartdata/mysql-operator-sidecar-8.0 \
  --set sidecar80.image.tag=v0.7.5-1 \
  --wait --timeout 5m
