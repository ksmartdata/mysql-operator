#!/bin/sh
# Usage: wait-node-conditions.sh <pod>=<condition-type>...
#   e.g. wait-node-conditions.sh e2e-mysql-0=Master e2e-mysql-1=Replicating
# Polls status.nodes of mysqlcluster/e2e (the operator<->orchestrator API
# contract) until every <pod>'s <condition-type> is True.
# WAIT_ITERS overrides the number of iterations (default 60, 5s apart).
set -eu

iters="${WAIT_ITERS:-60}"
i=1
while [ "$i" -le "$iters" ]; do
  status=$(kubectl get mysqlcluster e2e -n "$NAMESPACE" -o json 2>/dev/null) || status=""
  if [ -n "$status" ]; then
    all_ok=1
    for pair in "$@"; do
      pod=${pair%%=*}
      cond=${pair#*=}
      v=$(printf '%s' "$status" | jq -r --arg n "$pod.mysql.$NAMESPACE" --arg c "$cond" \
        '.status.nodes[]? | select(.name==$n) | .conditions[]? | select(.type==$c) | .status')
      if [ "$v" != "True" ]; then
        all_ok=0
        break
      fi
    done
    if [ "$all_ok" = "1" ]; then
      exit 0
    fi
  fi
  sleep 5
  i=$((i + 1))
done
echo "timed out waiting for node conditions: $*"
kubectl get mysqlcluster e2e -n "$NAMESPACE" -o jsonpath='{.status.nodes}' || true
exit 1
