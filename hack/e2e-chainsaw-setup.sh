#!/usr/bin/env bash
# Deploy the mysql-operator chart for e2e (shared by CI and local runs so the
# two copies of this command cannot drift apart)
set -euo pipefail

OPERATOR_IMAGE_REPO="${OPERATOR_IMAGE_REPO:-mysql-operator}"
OPERATOR_IMAGE_TAG="${OPERATOR_IMAGE_TAG:-e2e}"

# podSecurityContext=null matches the mcamel production chart: the default
# runAsUser 65532 makes the orchestrator container unable to write
# /etc/orchestrator/orchestrator.conf.json (CrashLoop)
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
