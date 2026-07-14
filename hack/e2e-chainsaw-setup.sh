#!/usr/bin/env bash
# Deploy the mysql-operator chart for e2e (shared by CI and local runs so the
# two copies of this command cannot drift apart)
set -euo pipefail

OPERATOR_IMAGE_REPO="${OPERATOR_IMAGE_REPO:-mysql-operator}"
OPERATOR_IMAGE_TAG="${OPERATOR_IMAGE_TAG:-e2e}"
# built from images/mysql-operator-orchestrator/Dockerfile in the same run,
# so the gate tests the orchestrator (Percona fork) shipped by the PR
ORCHESTRATOR_IMAGE_REPO="${ORCHESTRATOR_IMAGE_REPO:-mysql-operator-orchestrator}"
ORCHESTRATOR_IMAGE_TAG="${ORCHESTRATOR_IMAGE_TAG:-e2e}"
# CI overrides these with images that overlay the PR-built sidecar binary
# (hack/development/Dockerfile.sidecar-e2e); the defaults are the released
# images for local runs without a sidecar build
SIDECAR57_IMAGE_REPO="${SIDECAR57_IMAGE_REPO:-ghcr.io/ksmartdata/mysql-operator-sidecar-5.7}"
SIDECAR57_IMAGE_TAG="${SIDECAR57_IMAGE_TAG:-v0.7.4-1}"
SIDECAR80_IMAGE_REPO="${SIDECAR80_IMAGE_REPO:-ghcr.io/ksmartdata/mysql-operator-sidecar-8.0}"
SIDECAR80_IMAGE_TAG="${SIDECAR80_IMAGE_TAG:-v0.7.5-1}"

# podSecurityContext=null matches the mcamel production chart: the default
# runAsUser 65532 makes the orchestrator container unable to write
# /etc/orchestrator/orchestrator.conf.json (CrashLoop)
helm install mysql-operator ./deploy/charts/mysql-operator \
  --set podSecurityContext=null \
  --set image.repository="$OPERATOR_IMAGE_REPO" \
  --set image.tag="$OPERATOR_IMAGE_TAG" \
  --set image.pullPolicy=IfNotPresent \
  --set orchestrator.image.repository="$ORCHESTRATOR_IMAGE_REPO" \
  --set orchestrator.image.tag="$ORCHESTRATOR_IMAGE_TAG" \
  --set orchestrator.image.pullPolicy=IfNotPresent \
  --set sidecar57.image.repository="$SIDECAR57_IMAGE_REPO" \
  --set sidecar57.image.tag="$SIDECAR57_IMAGE_TAG" \
  --set sidecar80.image.repository="$SIDECAR80_IMAGE_REPO" \
  --set sidecar80.image.tag="$SIDECAR80_IMAGE_TAG" \
  --wait --timeout 5m
