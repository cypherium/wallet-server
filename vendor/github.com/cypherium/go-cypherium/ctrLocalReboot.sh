#!/bin/sh
noNumber=9
username="ubuntu"
port="22"
timeout=3
targetPath=/home/ubuntu
dateStr=$(date +%Y%m%d)
logName="cypherlog"
prjPath="gopath/src/github.com/cypherium/go-cypherium"
#pemFile="../../../../ansible-batch-control-aws/cypherium-us-west-key-pair.pem"
pemFile="./cypherium-us-west-key-pair.pem"
NetWorkId=`less ./genesis.json|awk -F "[:]" '/chainId/{print $2}'`
NetWorkId=`echo $NetWorkId | cut -d \, -f 1`
echo `pwd`
runShFile="runlan3.sh"
isMyMachine="0"
toolsPath="./cmd/tools"
for i in $(cat ./ctrLocalHostName)   #hosts文件里包含
do
    ip=$(echo "${i}" |awk -F":" '{print $1}')
    userName=$(echo "${i}" |awk -F":" '{print $2}')
    password=$(echo "${i}" |awk -F":" '{print $3}')
    echo "ip:$ip,userName:$userName,password:$password"
    echo $ip
    echo $userName
    echo "ssh"
    userPrjPath=/home/$userName/$prjPath
     case $ip in
     "192.168.0.153") runShFile="runlan1.sh"
     ;;
     "192.168.0.154") runShFile="runlan2.sh"
     ;;
     "192.168.0.153") runShFile="runlan3.sh"
     ;;
     *) runShFile="runlan3.sh"
     ;;
    esac

    ssh root@$ip  "sudo shutdown -r now;"
    #$targetRpcFile miner.start --local 1 --role 3 --ip $ip --preFix lan3
    #$targetRpcFile miner.start --local 1 --role 3 --ip 192.168.0.153 --preFix lan3
done