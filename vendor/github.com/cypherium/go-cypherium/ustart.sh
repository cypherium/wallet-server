#!/bin/sh
NODE_DIR="db"
GENESISDIR="./genesis.json"
BINDIR="./build/bin/cypher"
rm -rf db/
$BINDIR --datadir "$NODE_DIR" init $GENESISDIR
NetWorkId=`less genesis.json|awk -F "[:]" '/chainId/{print $2}'`
NetWorkId=`echo $NetWorkId | cut -d \, -f 1`
#sudo zerotier-cli join 83048a063200db6c
#kill -9 $(lsof -i:30301 |awk '{print $2}' | tail -n 2)
BOOTDIR="cnode://733edf2d1cb2faa843f71c54e0ab9fff676b6f3e937cefb44522d9acdfaf3a0208290a0b7d6aa58398412b95de86230c9ddfffe3af6e5e19455e7be9d9c458bd@[18.237.213.2]:30301"
$BINDIR  --onetport 7200 --nat "none" --ws --tps -wsaddr="0.0.0.0" --wsorigins "*" --rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcapi cph,web3,personal,miner,txpool,admin,net,txpool --port 14002 --rpcport 11002 --verbosity 3 --datadir db --networkid $NetWorkId --gcmode archive  --bootnodes $BOOTDIR > "$NODE_DIR/$n/$n.log" 2>&1 &
