#!/bin/sh
userPath=/root
userName="root"
passWord="1"
dateStr=$(date +%Y%m%d)
logName="cypherlog"
prjPath="work/src/github.com/cypherium/go-cypherium"
port="22"
timeout=3
logName="cypherlog"
logFileSize=2
if [[ "$2" != " " ]];then
 logFileSize=$2
fi
rm -rf awsLogs*
NetWorkId=`less ./genesis.json|awk -F "[:]" '/chainId/{print $2}'`
NetWorkId=`echo $NetWorkId | cut -d \, -f 1`
echo `pwd`

Do()
{
  host=$1
  echo $host
  userCphPrjPath=$userPath/$prjPath
  chaindbNodePath=$userCphPrjPath/chaindb/1
  result=""
  result=`ssh $userName@$host  "cd $chaindbNodePath; sudo rm -f $logName*.tar.bz2; sudo tar jcvf $logName.tar.bz2 ./output.log;"`
  echo $result
  scp  $userName@$host:$chaindbNodePath/$logName.tar.bz2 $logFileName/$host.tar.bz2
  mkdir -p $logFileName/$host
  sudo tar jxvf $logFileName/$host.tar.bz2 -C $logFileName/$host/
  cp $logFileName/$host/output.log $logFileName/$host.log
  rm -rf $logFileName/$host/
  rm -rf $logFileName/$host.tar.bz2
}
logFileName=awsLogs-$NetWorkId-$dateStr
if [ ! -d "$logFileName" ]; then
 mkdir -p $logFileName
fi



isAll=$1
if [[ "$isAll" == "0" ]];then
rm -rf $logFileName/*
for host in `sudo cat hostname.txt`;do
echo $host
  userCphPrjPath=$userPath/$prjPath
  chaindbNodePath=$userCphPrjPath/chaindb/1
result=""
result=`ssh $userName@$host  "cd $chaindbNodePath; sudo rm -f $logName*.tar.bz2; sudo tar jcvf $logName.tar.bz2 ./output.log;"`
echo $result
scp  $userName@$host:$chaindbNodePath/$logName.tar.bz2 $logFileName/$host.tar.bz2
mkdir -p $logFileName/$host
sudo tar jxvf $logFileName/$host.tar.bz2 -C $logFileName/$host/
#cp $logFileName/$host/output.log $logFileName/$host.log
split -b ${logFileSize}m $logFileName/$host/output.log $logFileName/$host/$host
#rm -rf $logFileName/$host/
rm -rf $logFileName/$host.tar.bz2
done
else
 Do $1
fi
echo "Done"
sudo chmod -R 777 $logFileName
