#!/usr/bin/env bash
noNumber=200
./cmd/tools/rpcctl genJsMap --local 1 --idxCg 1 --cnum $noNumber --preFix lan3 --sIdx 0
./cmd/tools/rpcctl genJsMap --local 1 --idxCg 1 --cnum $noNumber --preFix lan1 --sIdx 0
./cmd/tools/rpcctl genJsMap --local 1 --idxCg 1 --cnum $noNumber --preFix lan2 --sIdx 0
./cmd/tools/rpcctl genJsMap --local 1 --idxCg 1 --cnum $noNumber --preFix lan4 --sIdx 0
if [[ "$1" != "1" ]];then
./cmd/tools/rpcctl miner.start --role 3 --port 8000
./cmd/tools/rpcctl autoTrans --en 1 --time 10  --idx 4 --port 8000
#./cmd/tools/rpcctl miner.status
#./cmd/tools/rpcctl txBlockNumber
#./cmd/tools/rpcctl keyBlockNumber
#personal.unlockAll("1") cph.autoTransaction(1, 1)
else

if [[ "$2" != " " ]];then
  noNumber=$2
fi
./cmd/tools/rpcctl miner.start --role 3 --local 1 --port 18002 --preFix lan3
./cmd/tools/rpcctl miner.start --role 3 --local 1 --port 18002 --preFix lan1
./cmd/tools/rpcctl miner.start --role 3 --local 1 --port 18002 --preFix lan2
./cmd/tools/rpcctl miner.start --role 3 --local 1 --port 18002 --preFix lan4
./cmd/tools/rpcctl autoTrans --en 1 --time 100  --idx 1 --local 1 --port 18002 --preFix lan3
fi







