#!/bin/sh
genesisFile=./genesisLocal.json
dateStr=$(date +%Y%m%d)
prjPath="gopath/src/github.com/cypherium/go-cypherium/"
NetWorkId=`less $genesisFile|awk -F "[:]" '/chainId/{print $2}'`
NetWorkId=`echo $NetWorkId | cut -d \, -f 1`
nodeNum=9
echo `pwd`
logFileName="Lenovo-$NetWorkId-$dateStr"
logFileSize=2
if [[ "$2" != " " ]];then
 logFileSize=$2
fi
if [ ! -d "$logFileName" ]; then
 mkdir -p $logFileName
fi

rm -rf $logFileName/*

for i in $(cat ./ctrLocalHostName);do  #hosts文件里包含
    ip=$(echo "${i}" |awk -F":" '{print $1}')
    userName=$(echo "${i}" |awk -F":" '{print $2}')
    password=$(echo "${i}" |awk -F":" '{print $3}')
    echo "ip:$ip,host:$userName,password:$password"

    if [ ! -d "$logFileName/$ip-$userName" ]; then
     mkdir -p $logFileName/$ip-$userName
    fi
     #ssh  $userName@$ip 'bash -s' <./LenovoCopy.sh.sh $userHomePath $logName
     userHomePath=/home/$userName/
     scp $userName@$ip:$userHomePath/*.log $logFileName/$ip-$userName/
     #split -b ${logFileSize}m $logFileName/$host/output.log $logFileName/$host/$host

done

echo "Done"
sudo chmod -R 777 $logFileName






