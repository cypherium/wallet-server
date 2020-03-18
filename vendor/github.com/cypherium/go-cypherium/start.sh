#!/bin/bash

systemctl stop cypher.service
sleep .5

if [[ "$1" == "pull" ]];then
    git pull origin reconfigTestNet
    make cypher
fi

systemctl start cypher.service
sleep .1

systemctl status cypher.service -l
sleep 2

./build/bin/cypher attach chaindb/1/cypher.ipc

