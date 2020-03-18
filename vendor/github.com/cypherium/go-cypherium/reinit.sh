#!/bin/bash
CLEANDB="cldbf"
cleanDbCmd=$1
sudo chmod -R 777 ./run.sh
sudo chmod -R 777 ./genesis.json
sudo chmod -R 777 ./attach.sh
git pull
make cypher
make bootnode
systemctl stop cypher.service
sleep 2
rm -rf ./chaindb/1/output.log
#rm -rf /root/.cphash/*
if [[ "$cleanDbCmd" == "$CLEANDB" ]];then
  sudo ./run.sh kill 1
  sudo ./run.sh cldbf
  ./run.sh init 1
fi
rm -rf /root/work/tcpdump*
./install-tcpdmp.sh
sudo systemctl start cypher.service
sleep .5
sudo systemctl status cypher.service -l
sleep .5
sudo systemctl start tcpdumpsvc
sleep .5
sudo systemctl status tcpdumpsvc

