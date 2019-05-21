#!/bin/bash -e

TARGET="$target"

IFS=' ' read -r -a PARAM_HOST_ARRAY <<< "$1"

DEPLOY_DIR="/opt/framework"

ENV=${PARAM_HOST_ARRAY[0]}
HOST=${PARAM_HOST_ARRAY[1]}

function deploy_filer() {
    RSYNC_APP_PARAM="--include=config/*** --include=main --exclude=*"
    APP_NAME="filer"
    APP_DIR="$DEPLOY_DIR/$APP_NAME"
    CURRENT_DIR=$APP_DIR/current
    echo "[1m[32mPackage: $APP_NAME[0m"
    # TODO(benjamin): build app

    echo "[1m[32mRsync: $APP_NAME[0m"
    rsync -rpcv --chmod=Du=rwx,Dgo=rx,Fu=rw,Fgo=r --delete $RSYNC_APP_PARAM $HOST:$CURRENT_DIR
}

function deploy_task() {
    echo "TODO"
}
if [[ "$TARGET" == *"filer"* ]]; then
  deploy_filer
fi

if [[ "$TARGET" == *"task"* ]]; then
  deploy_task
fi
