#!/bin/sh
port="22"
timeout=3
userPath=/root
userName="root"
passWord="1"
dateStr=$(date +%Y%m%d)
logName="cypherlog"
prjPath="work/src/github.com/cypherium/go-cypherium"
NetWorkId=`less ./genesis.json|awk -F "[:]" '/chainId/{print $2}'`
NetWorkId=`echo $NetWorkId | cut -d \, -f 1`
echo `pwd`
runShFile="run.sh"
toolsPath="./cmd/tools"
branch="dTN-0.2"
CLEANDB="cldbf"

gitbranch()
{
   br=`git branch | grep "*"`
   echo ${br/* /}
}

branch=$(gitbranch)
echo "branch:$branch"

cleanDbCmd=$1
cleanAllDb()
{
for i in $(cat ./hostname.txt)   #hosts文件里包含
do
    ip=$(echo "${i}" |awk -F":" '{print $1}')
    echo "ip:$ip,userName:$userName,password:$passWord"
    echo $ip
    echo $userName
    echo "ssh"
    userCphPrjPath=$userPath/$prjPath
    echo "userCphPrjPath:$userCphPrjPath"
    ssh -i ./gcp_cypherium_private.pem $userName@$ip  "cd $userCphPrjPath;git reset --hard;git clean -f;rm -rf $userCphPrjPath/chaindb/cypher;rm -rf $userCphPrjPath/chaindb/17;git stash;$userCphPrjPath/$runShFile kill;$userCphPrjPath/$runShFile cldbf 1;exit;"
done
}
ostype()
{
  osname=`uname -s`
 # echo "osname $osname"
  rpcFileName="rpcctl"
  case $osname in

     "Linux") rpcFileName="linuxRpcCtl"
     ;;
     "Darwin") rpcFileName="rpcctl"
     ;;
     "linux") rpcFileName="linuxRpcCtl"
     ;;
     "darwin") rpcFileName="rpcctl"
     ;;
     *) rpcFileName="rpcctl"
     ;;
    esac
  return 0
}

ostype
targetRpcFile=$toolsPath/$rpcFileName
echo "cleanDbCmd:$cleanDbCmd"
if [[ "$cleanDbCmd" == "cldbf" ]];then
cleanAllDb
fi

for i in $(cat ./hostname.txt)   #hosts文件里包含
do
    ip=$(echo "${i}" |awk -F":" '{print $1}')
    echo "ip:$ip,userName:$userName,password:$passWord"
    echo $ip
    echo $userName
    echo "ssh"
    userCphPrjPath=$userPath/$prjPath
    echo "userCphPrjPath:$userCphPrjPath"
    ssh -i ./gcp_cypherium_private.pem $userName@$ip  "source /etc/profile;cd $userCphPrjPath;git reset --hard;git clean -f;git pull;git checkout $branch -f;git pull;cp -rf crypto/bls/lib/linux/* crypto/bls/lib/;make cypher;$userCphPrjPath/$runShFile kill;killall -9 cypher;$userCphPrjPath/reinit.sh $cleanDbCmd;exit;"

done
    sleep .5
   $targetRpcFile miner.start  --role 3 --passwd ZSRgWhO5%j
   # $targetRpcFile peerCount   --preFix lan1
    sleep 1
   $targetRpcFile autoTrans  --en 1 --time 10 --idx 1 --passwd ZSRgWhO5%j
   $targetRpcFile autoTrans  --en 1 --time 10 --idx 2 --passwd ZSRgWhO5%j
   # $targetRpcFile autoTrans  --en 1 --time 1 --preFix lan1 --idx 1
    #$targetRpcFile autoTrans  --en 1 --time 1 --preFix lan2 --idx 1 
    #$targetRpcFile miner.status   --preFix lan1 
    #$targetRpcFile txBlockNumber   --preFix lan1 
    #$targetRpcFile keyBlockNumber   --preFix lan1 
   # $targetRpcFile txpool.status   --preFix lan1 
