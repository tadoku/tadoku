#!/bin/bash

send_discord_notification() {
  curl --request POST \
    --url "$DISCORD_WEBHOOK_URL" \
    --header 'Content-Type: application/json' \
    --data '{ "content": "'"$1"'" }'
}

send_failure_notification() {
  send_discord_notification ":red_circle: ${PRODUCT_NAME} database backup failed, needs investigation"
  exit 1
}

trap 'send_failure_notification' ERR

export DUMP_FILE=/backup_${PRODUCT_NAME}_$(date +%Y%m%d_%H%M%S).pgdump
PGPASSWORD=$POSTGRES_PASSWORD pg_dump -d "$POSTGRES_DB" -U "$POSTGRES_USER" -h "$POSTGRES_HOST" -f "$DUMP_FILE" --inserts --no-owner --schema=$SCHEMA

# debugging
if [ "$DEBUG" == 'true' ]
then
  ls -al /

  echo "POSTGRES_DB: ${POSTGRES_DB}"
  echo "POSTGRES_USER: ${POSTGRES_USER}"
  echo "AWS_DEFAULT_REGION: ${AWS_DEFAULT_REGION}"
  echo "S3_BACKUP_PATH: ${S3_BACKUP_PATH}"
  echo "DUMP_FILE: ${DUMP_FILE}"

  cat "${DUMP_FILE}"

  echo "aws s3 cp ${DUMP_FILE}.bz2 ${S3_BACKUP_PATH} --storage-class ${STORAGE_CLASS}"
fi

bzip2 "$DUMP_FILE"
# storage class options: STANDARD | REDUCED_REDUNDANCY | STANDARD_IA | ONEZONE_IA | INTELLIGENT_TIERING | GLACIER | DEEP_ARCHIVE | OUTPOSTS | GLACIER_IR
aws s3 cp "${DUMP_FILE}".bz2 "$S3_BACKUP_PATH" --storage-class ${STORAGE_CLASS}

send_discord_notification ":green_circle: ${PRODUCT_NAME} database backup ran successfully"
