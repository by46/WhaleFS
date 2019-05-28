#!/bin/bash -e

REPO_DIR=$WORKSPACE
TARGET="$target"
ENV=${PARAM_HOST_ARRAY[0]}

# golang build environment
export HOME=/root
export PATH=/opt/go1.12/bin/:$PATH

ansible all -i "localhost," -m git -a "repo=git@cmisgitlab01:framework/yzw-playbooks.git dest=./yzw-playbooks/ version=develop" --connection=local

cd ./yzw-playbooks

ansible-playbook playbooks/whalefs/deploy.yml -e app_version=develop -e env=$ENV -e target=$TARGET -e need_checkout_source=false -e repo_dir=$REPO_DIR -e repo_name="" -t deploy
