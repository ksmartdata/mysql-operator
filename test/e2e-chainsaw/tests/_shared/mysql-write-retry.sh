#!/bin/sh
# 用法: mysql-write-retry.sh <pod> <id> <val>
# 以 sys_operator 账号向 <pod> 的 sys_operator.e2e_t 写入 (id, val)。
# 写入带重试：Ready 刚过时 pod 可能处于 operator 触发的瞬时滚动中。
set -eu

pod=$1; id=$2; val=$3
. "$(dirname "$0")/lib.sh"
fetch_oppass

err=""
i=1
while [ "$i" -le 24 ]; do
  if err=$(kubectl exec -n "$NAMESPACE" "$pod" -c mysql -- \
    mysql --no-defaults -h127.0.0.1 -usys_operator -p"$OPPASS" -e \
    "CREATE TABLE IF NOT EXISTS sys_operator.e2e_t (id INT PRIMARY KEY, val VARCHAR(64)); REPLACE INTO sys_operator.e2e_t VALUES ($id, '$val');" 2>&1); then
    exit 0
  fi
  echo "写入尝试 $i 失败: $err"
  sleep 5
  i=$((i + 1))
done
echo "写入失败: $pod ($id, '$val')"
kubectl exec -n "$NAMESPACE" "$pod" -c mysql -- \
  mysql --no-defaults -h127.0.0.1 -usys_operator -p"$OPPASS" -e \
  "SELECT @@hostname, @@read_only, @@super_read_only" 2>&1 || true
exit 1
