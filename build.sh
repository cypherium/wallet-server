#!/usr/bin/env bash

#set -x
appname='scan'

buildOnDarwin()
{
  cp -rf vendor/github.com/cypherium/cypherBFT/go-cypherium/crypto/bls/lib/mac/* vendor/github.com/cypherium/cypherBFT/go-cypherium/crypto/bls/lib/
  go build  -o ./bin/$appname ./src/main.go && (echo "BUILD SUCCESS"; exit 0;) || (echo "BUILD FAILED" && exit 1); 
}

buildOnLinux()
{
 cp -rf vendor/github.com/cypherium/cypherBFT/go-cypherium/crypto/bls/lib/linux/* vendor/github.com/cypherium/cypherBFT/go-cypherium/crypto/bls/lib/
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build  -o ./bin/$appname ./src/main.go && (echo "BUILD SUCCESS"; exit 0;) || (echo "BUILD FAILED" && exit 1);
}
ostype()
{
  osname=`uname -s`
  echo "osname $osname"
  echo "start build ..."
  case $osname in
     "Linux") buildOnLinux;
     ;;
     "Darwin")  buildOnDarwin;
     ;;
     "linux") buildOnLinux;
     ;;
     "darwin")  buildOnDarwin;
     ;;
     *) buildOnLinux;
     ;;
    esac
  return 0
}
rm -rf ./bin/$appname
ostype
###  build      ####
#echo "start build ..."
#Mac
# go build  -o ./bin/$appname ./src/main.go && (echo "BUILD SUCCESS"; exit 0;) || (echo "BUILD FAILED" && exit 1);
#linux
#CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build  -o ./bin/$appname ./src/main.go && (echo "BUILD SUCCESS"; exit 0;) || (echo "BUILD FAILED" && exit 1);
# xgo -x -targets=linux/amd64  ./src/ && (mv src-linux-amd64 ./bin/$appname; echo "BUILD SUCCESS"; exit 0;) || (echo "BUILD FAILED" && exit 1) || exit 1;

dir=$(pwd)
echo "initalize ..."
echo "rm $dir/output"
rm -rf $dir/output/ 2>&1 > /dev/null
mkdir -p  output/bin
mkdir -p  output/conf
mkdir -p  output/log
mkdir -p  output/logs

### copy files  ####
echo "copy to destination dir"
cp -R ./bin/scan                  output/bin/$appname
cp ./conf/scan.conf     ./output/conf/


### shell script ####
cp ./load.sh ./output/

### tar ############
echo "tar ..."
cd output
tar -czf $appname.tar.gz ./bin ./conf  ./log ./load.sh
mv ./$appname.tar.gz $dir/
rm -rf $dir/log/*

./load.sh restart


