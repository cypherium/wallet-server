#!/usr/bin/env bash

NODES_NUM=$1
PREFIX="local"
if [[ "$2" != " " ]];then
  PREFIX=$2
fi

[ "$#" -eq 0 ] && NODES_NUM=1

if ! [[ "$NODES_NUM" =~ ^[1-9][0-9]*$ ]];then
    echo "Invalid nodes number."
    exit 1
fi

ATTACH="${PREFIX}Chaindb/$NODES_NUM/cypher.ipc"

echo "Attach to $ATTACH"
./build/bin/cypher attach "$ATTACH"