#!/bin/bash -e

TARGET="$target"

IFS=' ' read -r -a PARAM_HOST_ARRAY <<< "$1"

DEPLOY_DIR="/opt/framework"

ENV=${PARAM_HOST_ARRAY[0]}
HOST=${PARAM_HOST_ARRAY[1]}

# golang build environment
export HOME=/root
export PATH=/opt/go1.12/bin/:$PATH

make build


function deploy_filer() {
    APP_NAME="filer"
    APP_DIR="$DEPLOY_DIR/$APP_NAME"
    CURRENT_DIR=$APP_DIR/current

    echo "[1m[32mRsync: $APP_NAME[0m"
    rsync -rpcv --chmod=Du=rwx,Dgo=rx,Fu=rw,Fgo=r --delete dist/ $HOST:$CURRENT_DIR
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
