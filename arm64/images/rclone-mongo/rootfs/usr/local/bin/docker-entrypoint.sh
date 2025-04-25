#!/bin/bash

set -x
set -o errexit

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

# exec command
case "$1" in
    delete)
        # delete
        echo "Delete s3:${STORE_PATH}"
        exec rclone --config=/tmp/rclone.conf deletefile s3:${STORE_PATH}
        ;;
    backup)
        # backup
        echo "Backup from mongo [${MONGO_HOST}]"
        mongodump -h ${MONGO_HOST} -u ${MONGO_USERNAME} -p ${MONGO_PASSWORD} --archive=/tmp/${BACKUP_FILE_NAME} --gzip
        echo "Copy to s3:${STORE_PATH}"
        exec rclone --config=/tmp/rclone.conf copy /tmp/${BACKUP_FILE_NAME} s3:${STORE_PATH}
        ;;
    recover)
        # recover
        echo "Copy from s3:${STORE_PATH} to /tmp"
        # get final element of path
        final_store_path=$(echo ${STORE_PATH} | awk -F'/' '{print $NF}')
        rclone --config=/tmp/rclone.conf copy s3:${STORE_PATH} /tmp
        echo "Restore to MongoDB [${MONGO_HOST}]"
        mongorestore -h ${MONGO_HOST} -u ${MONGO_USERNAME} -p ${MONGO_PASSWORD} --authenticationDatabase admin --archive=/tmp/${final_store_path} --gzip
        ;;
    *)
        echo "Usage: $0 {backup|recover|delete}"
        echo "Now runs your command."
        echo "$@"

        exec "$@"
esac