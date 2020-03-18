#!/bin/sh
userPath=/root
userName="root"
passWord="1"
dateStr=$(date +%Y%m%d)
logName="tcpdump"
prjPath="work"
folderName="tcpdumpLogs"
port="22"
timeout=3
logFileSize=2
if [[ "$2" != " " ]];then
 logFileSize=$2
fi
rm -rf $folderName*

Do()
{
  host=$1
  echo $host
  userCphPrjPath=$userPath/$prjPath
  tcpdumpPath=$userCphPrjPath
  result=""
  result=`ssh -i ./gcp_cypherium_private.pem $userName@$host  "cd $tcpdumpPath; sudo rm -f $logName*.tar.bz2; sudo tar jcvf $logName.tar.bz2 ./$logName*;"`
  echo $result
  scp -i ./gcp_cypherium_private.pem  $userName@$host:$tcpdumpPath/$logName.tar.bz2 $logFileName/$host.tar.bz2
  mkdir -p $logFileName/$host
  sudo tar jxvf $logFileName/$host.tar.bz2 -C $logFileName/$host/
  cp $logFileName/$host/output.log $logFileName/$host.log
  rm -rf $logFileName/$host/
  rm -rf $logFileName/$host.tar.bz2
}
logFileName=folderName
if [ ! -d "$logFileName" ]; then
 mkdir -p $logFileName
fi



isAll=$1
if [[ "$isAll" == "0" ]];then
rm -rf $logFileName/*
for host in `sudo cat hostname.txt`;do
echo $host
  userCphPrjPath=$userPath/$prjPath
  tcpdumpPath=$userCphPrjPath/chaindb/1
result=""
result=`ssh -i ./gcp_cypherium_private.pem $userName@$host  "cd $tcpdumpPath; sudo rm -f $logName*.tar.bz2; sudo tar jcvf $logName.tar.bz2 ./output.log;"`
echo $result
scp -i ./gcp_cypherium_private.pem  $userName@$host:$tcpdumpPath/$logName.tar.bz2 $logFileName/$host.tar.bz2
mkdir -p $logFileName/$host
sudo tar jxvf $logFileName/$host.tar.bz2 -C $logFileName/$host/
#cp $logFileName/$host/output.log $logFileName/$host.log
#rm -rf $logFileName/$host/
rm -rf $logFileName/$host.tar.bz2
done
else
 Do $1
fi
echo "Done"
sudo chmod -R 777 $logFileName
