#!/bin/bash
row=1
prjPath="/root/work/src/github.com/cypherium/go-cypherium"
echo $prjPath
NODE_DIR="$prjPath/chaindb"

ip=`curl icanhazip.com`
for i in $(cat $prjPath/hostname.txt)   #hosts文件里包含
do
    ipTemp=$(echo "${i}" |awk -F":" '{print $1}')
    if [[ "$ipTemp" == "$ip" ]];then
    break
    fi
    let row+=1
done

$prjPath/build/bin/cypher attach "$NODE_DIR/$row/cypher.ipc"
