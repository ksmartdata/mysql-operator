#!/bin/sh
# 用法: wait-node-conditions.sh <pod>=<条件类型>...
#   例: wait-node-conditions.sh e2e-mysql-0=Master e2e-mysql-1=Replicating
# 轮询 mysqlcluster/e2e 的 status.nodes（operator↔orchestrator API 契约），
# 直到所有 <pod> 的 <条件类型> 均为 True。WAIT_ITERS 覆盖轮询次数（默认 60，间隔 5s）。
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
echo "节点 condition 超时: $*"
kubectl get mysqlcluster e2e -n "$NAMESPACE" -o jsonpath='{.status.nodes}' || true
exit 1
