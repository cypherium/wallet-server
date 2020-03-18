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
gitBranch="dTN-0.3"
isMyMachine="0"
toolsPath="./cmd/tools"
subFix="lan1"

gitbranch()
{
   br=`git branch | grep "*"`
   echo ${br/* /}
}

gitBranch=$(gitbranch)
echo "branch:$branch"

selectRunShFile()
{
     ip=$1
    # selectShFile="runlan3.sh"
      case $ip in
     "192.168.0.153") subFix="lan1"
     ;;
     "192.168.0.154") subFix="lan2"
     ;;
     "192.168.0.168") subFix="lan3"
     ;;
     "192.168.0.115") subFix="lan4"
     ;;
     *) subFix="lan3"
     ;;
    esac
    selectShFile="run"$subFix".sh"
    runShFile=$selectShFile
   # return selectShFile
}

cleanAllNode()
{
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
    selectRunShFile $ip
    ssh $userName@$ip  "source ~/.bash_profile;cd $userPrjPath;git clean -f;git pull;git checkout $gitBranch -f;git pull;cp -rf crypto/bls/lib/linux/* crypto/bls/lib/;cp -rf ../jdk ./;make cypher;$userPrjPath/$runShFile kill;$userPrjPath/runlan1.sh cleandb 20;$userPrjPath/runlan2.sh cleandb 20;$userPrjPath/runlan3.sh cleandb 20;$userPrjPath/runlan4.sh cleandb 20;exit;"
    #$targetRpcFile miner.start --local 1 --role 3 --ip $ip --preFix lan3
    #$targetRpcFile miner.start --local 1 --role 3 --ip 192.168.0.153 --preFix lan3
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

isMyMachine=$1
if [[ "$isMyMachine" == "0" ]];then

cleanAllNode
$targetRpcFile genJsMap --local 1 --idxCg 1 --cnum $noNumber --preFix lan3 --sIdx 0
$targetRpcFile genJsMap --local 1 --idxCg 1 --cnum $noNumber --preFix lan1 --sIdx 0
$targetRpcFile genJsMap --local 1 --idxCg 1 --cnum $noNumber --preFix lan2 --sIdx 0

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
    selectRunShFile $ip

    ssh $userName@$ip  "source ~/.bash_profile;cd $userPrjPath;git reset --hard;git clean -f;git pull;git checkout $gitBranch -f;git pull;cp -rf crypto/bls/lib/linux/* crypto/bls/lib/;cp -rf ../jdk ./;make cypher;$userPrjPath/$runShFile init $noNumber;$userPrjPath/$runShFile $noNumber;exit;"
    #$targetRpcFile miner.start --local 1 --role 3 --ip $ip --preFix lan3
    #$targetRpcFile miner.start --local 1 --role 3 --ip 192.168.0.153 --preFix lan3
done
     sleep .5
    $targetRpcFile miner.start --local 1 --role 3  --preFix lan3 --port 18002
    $targetRpcFile miner.start --local 1 --role 3  --preFix lan1 --port 18002
    $targetRpcFile miner.start --local 1 --role 3  --preFix lan2 --port 18002
   # $targetRpcFile peerCount  --local 1 --preFix lan1
    sleep 1
    $targetRpcFile autoTrans --local 1 --en 1 --time 1 --preFix lan3 --idx 1 --port 18002
   # $targetRpcFile autoTrans --local 1 --en 1 --time 1 --preFix lan1 --idx 1 --port 18002
    #$targetRpcFile autoTrans --local 1 --en 1 --time 1 --preFix lan2 --idx 1 --port 18002
   # $targetRpcFile miner.status  --local 1 --preFix lan1 --port 18002
    $targetRpcFile txBlockNumber  --local 1 --preFix lan1 --port 18002
    #$targetRpcFile keyBlockNumber  --local 1 --preFix lan1 --port 18002
   # $targetRpcFile txpool.status  --local 1 --preFix lan1 --port 18002
else

$cleanAllNode
for i in $(cat ./ctrLocalHostName)   #hosts文件里包含
do
    ip=$(echo "${i}" |awk -F":" '{print $1}')
    if [[ "$ip" == "$1" ]];then
    userName=$(echo "${i}" |awk -F":" '{print $2}')
    password=$(echo "${i}" |awk -F":" '{print $3}')
    echo "ip:$ip,userName:$userName,password:$password"
    break
    fi
done

    ip=$(echo "${i}" |awk -F":" '{print $1}')
    userName=$(echo "${i}" |awk -F":" '{print $2}')
    password=$(echo "${i}" |awk -F":" '{print $3}')
    echo "ip:$ip,userName:$userName,password:$password"
    echo $ip
    echo $userName
    echo "ssh"
    userPrjPath=/home/$userName/$prjPath
    selectRunShFile $ip

   # runShFile="run"$subFix".sh"
    ssh $userName@$ip  "source ~/.bash_profile;cd $userPrjPath;git reset --hard;git clean -f;git pull;git checkout $gitBranch -f;git pull;cp -rf crypto/bls/lib/linux/* crypto/bls/lib/;cp -rf ../jdk ./;make cypher;$userPrjPath/$runShFile kill;$userPrjPath/$runShFile clnode $noNumber;$userPrjPath/$runShFile init $noNumber;$userPrjPath/$runShFile $noNumber;"

    sleep .5
    $targetRpcFile miner.start --local 1 --role 3  --preFix $subFix --port 18002
   # $targetRpcFile peerCount  --local 1 --preFix lan1
    sleep 1
    $targetRpcFile autoTrans --local 1 --en 1 --time 1 --preFix $subFix --idx 1 --port 18002
    $targetRpcFile autoTrans --local 1 --en 1 --time 1 --preFix $subFix --idx 2 --port 18002
    $targetRpcFile autoTrans --local 1 --en 1 --time 1 --preFix $subFix --idx 3 --port 18002
    $targetRpcFile miner.status  --local 1 --preFix $subFix --port 18002
    $targetRpcFile txBlockNumber  --local 1 --preFix $subFix --port 18002
    $targetRpcFile keyBlockNumber  --local 1 --preFix $subFix --port 18002
   # $targetRpcFile txpool.status  --local 1 --preFix lan1 --port 18002

fi




