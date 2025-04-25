#!/bin/bash

set -xe

echo "Downloading backup from S3..."
rclone config create s3 s3 endpoint ${S3_ENDPOINT} access_key_id ${S3_ACCESS_KEY_ID} secret_access_key ${S3_SECRET_ACCESS_KEY}
rclone copy s3:${BAK_FILE_PATH} /
gunzip /$(basename $BAK_FILE_PATH)
echo "Restore backup to database..."
export PGPASSWORD=$PG_PASS
psql -h $PG_HOST -p $PG_PORT -U $PG_USER < /$(basename $BAK_FILE_PATH .gz)

echo "Restore completed!"