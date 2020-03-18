#!/bin/bash

NODES_NUM=$1
[ "$#" -eq 0 ] && NODES_NUM=1

if ! [[ "$NODES_NUM" =~ ^[1-9][0-9]*$ ]];then
    echo "Invalid nodes number."
    exit 1
fi

ATTACH="localChaindb/$NODES_NUM/cypher.ipc"

echo "Attach to $ATTACH"
./build/bin/cypher attach "$ATTACH"