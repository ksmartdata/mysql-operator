#!/bin/sh
# Usage: mysql-read-until.sh <pod> <id> <expected>
# Polls <pod> until sys_operator.e2e_t holds val == expected for the given id.
# READ_ITERS / READ_SLEEP override the polling parameters (default 30 x 2s).
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
echo "read timed out: $pod never returned e2e_t id=$id val='$expected' (last read: '$out')"
exit 1
