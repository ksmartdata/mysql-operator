#!/bin/sh
# Usage: mysql-write-retry.sh <pod> <id> <val>
# Writes (id, val) into sys_operator.e2e_t on <pod> as the sys_operator user.
# Writes are retried: right after Ready the pod may be in a transient rollout
# triggered by the operator.
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
  echo "write attempt $i failed: $err"
  sleep 5
  i=$((i + 1))
done
echo "write failed: $pod ($id, '$val')"
kubectl exec -n "$NAMESPACE" "$pod" -c mysql -- \
  mysql --no-defaults -h127.0.0.1 -usys_operator -p"$OPPASS" -e \
  "SELECT @@hostname, @@read_only, @@super_read_only" 2>&1 || true
exit 1
