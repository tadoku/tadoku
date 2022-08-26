#!/bin/bash

export DUMP_FILE=/backup_${PRODUCT_NAME}_`date +%Y%m%d_%H%M%S`.pgdump
PGPASSWORD=$POSTGRES_PASSWORD pg_dump -d $POSTGRES_DB -U $POSTGRES_USER -h $POSTGRES_HOST -f $DUMP_FILE

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

  echo "aws s3 cp ${DUMP_FILE}.bz2 ${S3_BACKUP_PATH}"
fi

bzip2 $DUMP_FILE
aws s3 cp ${DUMP_FILE}.bz2 $S3_BACKUP_PATH --storage-class GLACIER_IR
