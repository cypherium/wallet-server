#!/bin/bash

HOST=35.232.249.84
appname=scan
approot=/root/work/src/github.com/cypherium/wallet-scan

#./build.sh || exit 1
echo "start copy to remote"
scp -i vendor/github.com/cypherium/cypherBFT/go-cypherium/gcp_cypherium_private.pem scan.tar.gz root@$HOST:$approot
ssh -i vendor/github.com/cypherium/cypherBFT/go-cypherium/gcp_cypherium_private.pem root@$HOST "cd $approot && tar -xvzf scan.tar.gz && ./load.sh restart"

#
### done ############
echo "done,done,done"
