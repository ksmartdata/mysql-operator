#!/bin/sh
# 用法: mysql-read-until.sh <pod> <id> <expected>
# 轮询 <pod>，直到 sys_operator.e2e_t 中 id 对应的 val == expected。
# READ_ITERS / READ_SLEEP 覆盖轮询参数（默认 30 次 × 2s）。
set -eu

pod=$1; id=$2; expected=$3
. "$(dirname "$0")/lib.sh"
fetch_oppass

iters="${READ_ITERS:-30}"
interval="${READ_SLEEP:-2}"
out=""
i=1
while [ "$i" -le "$iters" ]; do
  out=$(kubectl exec -n "$NAMESPACE" "$pod" -c mysql -- \
    mysql --no-defaults -h127.0.0.1 -usys_operator -p"$OPPASS" -N -e \
    "SELECT val FROM sys_operator.e2e_t WHERE id=$id" 2>/dev/null || true)
  if [ "$out" = "$expected" ]; then
    exit 0
  fi
  sleep "$interval"
  i=$((i + 1))
done
echo "读取超时: $pod 未读到 e2e_t id=$id val='$expected'（最后读到: '$out'）"
exit 1
