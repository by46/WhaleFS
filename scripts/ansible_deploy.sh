#!/bin/bash -e

REPO_DIR=$WORKSPACE

ansible all -i "localhost," -m git -a "repo=git@cmisgitlab01:framework/yzw-playbooks.git dest=./yzw-playbooks/ version=develop" --connection=local

cd ./yzw-playbooks

ansible-playbook playbooks/whalefs/qa/deploy_filer.yml -e app_version=develop -e need_checkout_source=false -e repo_dir=$REPO_DIR -e repo_name="" -t deploy

