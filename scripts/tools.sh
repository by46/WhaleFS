#!/bin/bash

CURRENT_DIR=current

function backup() {
  WORK_DIR=$1
  KEEP_NUM=$2

  DEPLOY_DIR=$(date +%Y%m%d%H%M%S)

  cd $WORK_DIR
  cp -r ${CURRENT_DIR} ${DEPLOY_DIR}

  for app_dir in $(find . -type d -name "20*" | sort -nr | awk -v keep_num=$KEEP_NUM '{ if (NR > keep_num) print $1}'); 
  do
    echo "deleting $app_dir ..."
    rm -rf $app_dir
  done
}

function rollback() {
  WORK_DIR=$1
  BACK_NUM=$2

  cd $WORK_DIR

  for app_dir in $(find . -type d -name "20*" | sort -nr | awk -v back_num=$BACK_NUM '{ if (NR = back_num) print $1}'); 
  do
    echo "rollback $app_dir ..."
    
  done
}

case $1 in

  backup)
    backup $2 10
    ;;
  
  rollback)
    rollback $2 1
    ;;

esac