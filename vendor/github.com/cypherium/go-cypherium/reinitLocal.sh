#!/bin/bash

./runlocal.sh kill
./runlocal.sh cleandb "$1"
make cypher
make cypher
./runlocal.sh init "$1"
./runlocal.sh "$1"

#./cmd/tools/auto trans --port "18006"
#./cmd/tools/auto miner --port "18004,18006,18008,18010,18012,18014,18002,18016"