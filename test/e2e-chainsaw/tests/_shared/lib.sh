# 共享函数库。chainsaw script 的工作目录是所在测试目录，
# 用 . ../_shared/lib.sh（或脚本内 $(dirname "$0")/lib.sh）引入。
# POSIX sh 无 pipefail：管道里 kubectl 失败不会中止 set -e 脚本，
# 所以取密码后必须显式校验非空，否则空密码会把排障引向错误方向。

fetch_oppass() {
  _i=1
  while [ "$_i" -le 3 ]; do
    OPPASS=$(kubectl get secret e2e-mysql-operated -n "$NAMESPACE" \
      -o jsonpath='{.data.OPERATOR_PASSWORD}' | base64 -d) || OPPASS=""
    if [ -n "$OPPASS" ]; then
      return 0
    fi
    echo "获取 OPERATOR_PASSWORD 失败（第 $_i 次）" >&2
    _i=$((_i + 1))
    sleep 2
  done
  echo "e2e-mysql-operated secret 不可用或密码为空" >&2
  return 1
}
