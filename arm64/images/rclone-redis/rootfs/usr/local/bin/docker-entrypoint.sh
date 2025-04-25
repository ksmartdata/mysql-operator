#!/bin/bash

echo "Create Google Drive service-account.json file."
echo "${GDRIVE_SERVICE_ACCOUNT}" > /tmp/gdrive-service-account.json

echo "Create rclone.conf file."
cat <<EOF > /tmp/rclone.conf
[gd]
type = drive
scope = drive
service_account_file = /tmp/gdrive-service-account.json
client_id = ${GDRIVE_CLIENT_ID}
root_folder_id = ${GDRIVE_ROOT_FOLDER_ID}
impersonate = ${GDRIVE_IMPERSONATOR}

[s3]
type = s3
env_auth = true
provider = ${S3_PROVIDER:-"AWS"}
access_key_id = ${AWS_ACCESS_KEY_ID}
secret_access_key = ${AWS_SECRET_ACCESS_KEY:-$AWS_SECRET_KEY}
region = ${AWS_REGION:-"us-east-1"}
endpoint = ${S3_ENDPOINT}
acl = ${AWS_ACL}
storage_class = ${AWS_STORAGE_CLASS}
session_token = ${AWS_SESSION_TOKEN}
no_check_bucket = true

[gs]
type = google cloud storage
project_number = ${GCS_PROJECT_ID}
service_account_file = /tmp/google-credentials.json
object_acl = ${GCS_OBJECT_ACL}
bucket_acl = ${GCS_BUCKET_ACL}
location =  ${GCS_LOCATION}
storage_class = ${GCS_STORAGE_CLASS:-"MULTI_REGIONAL"}

[http]
type = http
url = ${HTTP_URL}

[azure]
type = azureblob
account = ${AZUREBLOB_ACCOUNT}
key = ${AZUREBLOB_KEY}
EOF

if [[ -n "${GCS_SERVICE_ACCOUNT_JSON_KEY:-}" ]]; then 
    echo "Create google-credentials.json file."
    cat <<EOF > /tmp/google-credentials.json
    ${GCS_SERVICE_ACCOUNT_JSON_KEY}
EOF
else 
    touch /tmp/google-credentials.json
fi

# exec command
case "$1" in
    delete)
        # delete
        echo "delete s3:${STORE_PATH}"
        exec rclone --config=/tmp/rclone.conf deletefile s3:${STORE_PATH}
        ;;
    backup)
        # backup
        echo "backup from redis [${READIS_HOST}]"
        redis-cli -h ${READIS_HOST} -p 6379 -a ${REDIS_PASSWORD} --rdb ${RDB_NAME}
        echo "copy to s3:${STORE_PATH}"
        exec rclone --config=/tmp/rclone.conf copy ${RDB_NAME} s3:${STORE_PATH}
        ;;
    recover)
        # recover
        if [ "`ls -A ${REDIS_DATA_DIR}`" = "" ]; then
          echo "${REDIS_DATA_DIR} is indeed empty"
          echo "copy from s3:${STORE_PATH} to ${REDIS_DATA_DIR}"
          exec rclone --config=/tmp/rclone.conf copy s3:${STORE_PATH} ${REDIS_DATA_DIR}
        else
          echo "${REDIS_DATA_DIR} is not empty!Will not copy!"
        fi
        ;;
    *)
        echo "Usage: $0 {backup|recover|delete}"
        echo "Now runs your command."
        echo "$@"

        exec "$@"
esac
