#!/bin/bash
username="root"
port="22"
userPath="/root"
targetBinPath="../../CypherTestBin"
dateStr=$(date +%Y%m%d)
binName="cypher"
pemFile="../../../../ansible-batch-control-aws/cypherium-us-west-key-pair.pem"
prjPath="work/src/github.com/cypherium/go-cypherium"
cp -rf ./crypto/bls/lib/mac/*     ./crypto/bls/lib/
make cypher
for i in $(cat ./hostname.txt)   #ips文件里包含
do
    ip=$(echo "${i}" |awk -F":" '{print $1}')
    echo "ip:$ip,userName:$userName,password:$passWord"
    echo $ip
    echo $userName
    echo "ssh"
    userCphPrjPath=$userPath/$prjPath
    echo "userCphPrjPath:$userCphPrjPath"
    break
done

cp -rf ./genesis.json $targetBinPath/genesis.json
cp -rf ./build/bin/cypher $targetBinPath/mac/cypher
cp -rf ./crypto/bls/lib/linux/*     ./crypto/bls/lib/
ssh  -i ./gcp_cypherium_private.pem $username@$ip  "cd $userCphPrjPath; sudo rm -f $binName*.tar.bz2; cd ./build/bin;sudo tar jcvf $binName.tar.bz2 ./cypher;"
scp  -i ./gcp_cypherium_private.pem $username@$ip:$userCphPrjPath/build/bin/cypher.tar.bz2 $targetBinPath/linux/cypher.tar.bz2
chmod -R 777 $targetBinPath/linux
mkdir -p $targetBinPath/linux/bin
sudo tar jxvf $targetBinPath/linux/cypher.tar.bz2 -C $targetBinPath/linux/bin/
cp -rf $targetBinPath/linux/bin/ $targetBinPath/linux
rm -rf $targetBinPath/linux/cypher.tar.bz2 $targetBinPath/linux/bin/
echo "Done"
