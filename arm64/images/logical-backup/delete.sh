#!/bin/bash

# enable unofficial bash strict mode
#
# Only support delete from s3

set -o errexit
set -o nounset
set -o pipefail
IFS=$'\n\t'

ALL_DB_SIZE_QUERY="select sum(pg_database_size(datname)::numeric) from pg_database;"
PG_BIN=$PG_DIR/$PG_VERSION/bin
DUMP_SIZE_COEFF=5
ERRORCOUNT=0

TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
KUBERNETES_SERVICE_PORT=${KUBERNETES_SERVICE_PORT:-443}
if [ "$KUBERNETES_SERVICE_HOST" != "${KUBERNETES_SERVICE_HOST#*[0-9].[0-9]}" ]; then
    echo "IPv4"
    K8S_API_URL=https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT/api/v1
elif [ "$KUBERNETES_SERVICE_HOST" != "${KUBERNETES_SERVICE_HOST#*:[0-9a-fA-F]}" ]; then
    echo "IPv6"
    K8S_API_URL=https://[$KUBERNETES_SERVICE_HOST]:$KUBERNETES_SERVICE_PORT/api/v1
elif [ -n "$KUBERNETES_SERVICE_HOST" ]; then
    echo "Hostname"
    K8S_API_URL=https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT/api/v1
else
  echo "KUBERNETES_SERVICE_HOST was not set"
fi
echo "API Endpoint: ${K8S_API_URL}"
CERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt

LOGICAL_BACKUP_PROVIDER=${LOGICAL_BACKUP_PROVIDER:="s3"}
LOGICAL_BACKUP_S3_RETENTION_TIME=${LOGICAL_BACKUP_S3_RETENTION_TIME:=""}

function aws_delete_object {
    args=(
      "--bucket=$LOGICAL_BACKUP_S3_BUCKET"
    )
    # must LOGICAL_BACKUP_FILE_NAME not be empty
    if [[ -z "$LOGICAL_BACKUP_FILE_NAME" ]]; then
        echo "LOGICAL_BACKUP_FILE_NAME is empty"
        exit 1
    fi
    PATH_TO_BACKUP="/spilo/"$SCOPE$LOGICAL_BACKUP_S3_BUCKET_SCOPE_SUFFIX"/logical_backups/"$LOGICAL_BACKUP_FILE_NAME.sql.gz

    [[ ! -z "$LOGICAL_BACKUP_S3_ENDPOINT" ]] && args+=("--endpoint-url=$LOGICAL_BACKUP_S3_ENDPOINT")
    [[ ! -z "$LOGICAL_BACKUP_S3_REGION" ]] && args+=("--region=$LOGICAL_BACKUP_S3_REGION")

    aws s3api delete-object "${args[@]}" --key "${PATH_TO_BACKUP}"

    echo "Delete backup file ${PATH_TO_BACKUP} successfully"
}

aws_delete_object