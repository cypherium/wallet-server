## What is run.sh

run.sh is used for runs n local cypher quickly.

## Which version of the source code is available

```
$ git clone https://github.com/cypherium/cypherium/cypherBFT.git -b reconfiguration
```

## Running cypher locally
```
$ cd go-cypherium
$ make cypher
$ ./build/bin/cypher account new --datadir db/1 
password must be set 1. Copy 1 public key from terminal log to genesis.json
$ ./build/bin/cypher account new --datadir db/2
password must be set 1. Copy 2 public key from terminal log to genesis.json
$ ./run.sh init 2
$ cp -R cmd/cypher/jdk/ ./
$ ./run.sh 2
```
run.sh do the following things:
1./run.sh kill Kill all cypher processes if they exist in the background.
2./run.sh n Created n folders to save the data for each cypher,include blockchain db, output log info, private key,public key etc..
3.Bootstrap and initialize a new genesis txblock and keyblock for each cypher.
4.Runs n local cypher background.

## How to attach cypher
Open another console,enter the following command to attach one process,it will enter co1 javaScript interactive console. 

```
$ ./build/bin/cypher attach ./db/n/cypher.ipc
>cph.accounts
>miner.start(1, "address")
```
Now you can first use personal.newAccount('password') to create an account for recive rewards,then use bftcosi.start() which will run PBFT-COSI protocol as leader,while other processes as follower, all processes work together to generate txblock,finally, you can use bftcosi.stop() to terminated.

All processes log info saved in ./db/n/output.log.

