# Shared helpers. chainsaw scripts run with the test directory as the working
# directory; source this with . ../_shared/lib.sh (or $(dirname "$0")/lib.sh
# from within a script).
# POSIX sh has no pipefail: a kubectl failure inside a pipeline does not abort
# a set -e script, so the password must be explicitly checked for emptiness —
# otherwise an empty password sends debugging in the wrong direction.

fetch_oppass() {
  _i=1
  while [ "$_i" -le 3 ]; do
    OPPASS=$(kubectl get secret e2e-mysql-operated -n "$NAMESPACE" \
      -o jsonpath='{.data.OPERATOR_PASSWORD}' | base64 -d) || OPPASS=""
    if [ -n "$OPPASS" ]; then
      return 0
    fi
    echo "failed to fetch OPERATOR_PASSWORD (attempt $_i)" >&2
    _i=$((_i + 1))
    sleep 2
  done
  echo "e2e-mysql-operated secret unavailable or password empty" >&2
  return 1
}
