#!/bin/bash

TOP_PID=$$
CYPHER="./build/bin/cypher"
GENESIS="./genesis.json"
NODE_DIR="localChaindb"
NODES_MAX=10
NODES_NUM=0
ACCOUNTSLIST="$NODE_DIR/accounts.txt"
LOGLEVEL=6
NETWORKID=123663

# Commands
CLEANDB="cleandb"
INIT="init"
CLEAN="clean"
APPEND="append"
STOP="stop"
NEWACCOUNT="newAccount"
LISTACCOUNT="listAccount"
BOOTNODE="cnode://2d42a9f2ecdffeb165d4cc5e56aece74389de1c087132f131ec531b6627be2aeff97e001ee015ad08aa542c4aaee14052dd8e13c51e1f230dcc4c56e977edfea@54.202.212.33:30301"

log()
{
    echo 1>&2 "$1"
}

die()
{
    log "$1"
    kill -s TERM $TOP_PID
}

die_usage()
{
    log "Usage: $0 [init|cleandb] <number_of_nodes>"
    log "    init: init nodes from genesis file"
    log "    clean: remove nodes database including keystore"
    log "    cleandb: remove nodes' database except keystore"
    log "    stop: kill node by specify its number"
    log "    <number_of_nodes>: number of nodes you want to run"
    die
}

init_node()
{
    log "Init nodes...."
    for n in $( seq $NODES_NUM ); do
        mkdir -p "$NODE_DIR/$n"
        $CYPHER --datadir "$NODE_DIR/$n" init "$GENESIS"
    done
}


#
# List accounts for each nodes.
#
list_account()
{
    PASSWORD=`mktemp`
    echo '1' > $PASSWORD

    echo "" > "$ACCOUNTSLIST"
    for n in $( seq $NODES_NUM ); do
        printf "\n\nNode $n accounts:\n" >> "$ACCOUNTSLIST"
        $CYPHER account list --datadir "$NODE_DIR/$n" --password $PASSWORD >> "$ACCOUNTSLIST"
#        printf "$MSG" >> "$ACCOUNTSLIST"
    done

    rm -f $PASSWORD
    log "Check out $NODE_DIR/accounts.txt for accounts detail"
}


#
# Create 2 accounts for each nodes.
#
new_account()
{
    PASSWORD=`mktemp`
    echo '1' > $PASSWORD
    for n in $( seq $NODES_NUM ); do
        log "Create two accounts for node $n...."
        mkdir -p "$NODE_DIR/$n"
        $CYPHER account new --datadir "$NODE_DIR/$n" --password $PASSWORD
        $CYPHER account new --datadir "$NODE_DIR/$n" --password $PASSWORD
    done
    rm -f $PASSWORD

    list_account
}

#
# append_node: start additional nodes
# - argument 1: Nodes number already running
# - argument 2: additional nodes number to append
#
append_node()
{
    log "Append $NODES_NUM nodes...."
    for m in $( seq $NODES_NUM ); do
        n=$(($1 + m))
        log "Node $n -onetport@$((7100 + 2 * $n)) - port@$((16000 + 2 * $n)) - rpcport@$((18000 + 2 * $n)) - rpc path: $NODE_DIR/$n/cypher.ipc"
        # --ws  -wsaddr="0.0.0.0" --wsorigins "*"
        if [[ "$n" -eq 1 ]]; then
            $CYPHER  --onetport $((7100 + 2 * $n)) --nat "none" --ws  -wsaddr="0.0.0.0" --wsorigins "*" --tps --rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcapi cph,web3,personal,miner,txpool --port $((16000 + 2 * $n)) --rpcport $((18000 + 2 * $n)) --verbosity "$LOGLEVEL" --datadir "$NODE_DIR/$n" --networkid "$NETWORKID" --gcmode archive --bootnodes "$BOOTNODE" 2>"$NODE_DIR/$n/$n.log" &
        else
            $CYPHER  --onetport $((7100 + 2 * $n)) --nat "none" --tps --rpc --rpccorsdomain "*" --rpcaddr 0.0.0.0 --rpcapi cph,web3,personal,miner,txpool --port $((16000 + 2 * $n)) --rpcport $((18000 + 2 * $n)) --verbosity "$LOGLEVEL" --datadir "$NODE_DIR/$n" --networkid "$NETWORKID" --gcmode archive --bootnodes "$BOOTNODE" 2>"$NODE_DIR/$n/$n.log" &
        fi
    done
}

clean_db()
{
    log "Clean $NODES_NUM nodes' db, keystore is reserved."
    for n in $( seq $NODES_NUM ); do
        rm -rf $NODE_DIR/$n/cypher $NODE_DIR/$n/cypher.ipc $NODE_DIR/$n/output.log
    done
}

log_dangerous()
{
    log "dangerous"
}

#
# confirm_do: prompt user to input y/Y to confirm before doing something dangerous
# - argument 1: prompt information
# - argument 2: dangerous function
#
# example: confirm_do "Are you sure to delete database? " log_dangerous
#
confirm_do()
{
    read -p "$1" -n 2 -r
    if [[ $REPLY =~ ^[Yy]$ ]]
    then
        # do dangerous stuff
        "$2"
    fi
}

#--nat=extip:119.123.199.243
# parameters check

[ "$#" -eq 0 ] && die_usage

if [[ "$1" == "list" ]];then
    ps -a |grep "[r]pcaddr"
    exit 0
fi

if [[ "$1" == "kill" ]];then
    killall -HUP cypher
    exit 0
fi

rm_node()
{
    log "Clean all node data."
    rm -rf "$NODE_DIR"
}

if [[ "$1" == "$CLEAN" ]];then
    confirm_do "Are you sure to delete chain database including keystore?[y/N] " rm_node
    exit 0
fi

if ! [[ -x "$CYPHER" ]];then
    die "$CYPHER not found"
fi

if [[ "$1" == "$INIT"  || "$1" == "$CLEANDB" || "$1" == "$APPEND" || "$1" == "$STOP" || "$1" == "$NEWACCOUNT" || "$1" == "$LISTACCOUNT" ]];then
    NODES_NUM=$2
else
    NODES_NUM=$1
fi

if [[ "$1" == "$LISTACCOUNT" ]];then
    list_account
    exit 0
fi

if ! [[ "$NODES_NUM" =~ ^[1-9][0-9]*$ ]];then
    die "Invalid nodes number."
fi

if [[ "$NODES_NUM" -gt "$NODES_MAX" ]]; then
    log "Too many nodes to run, max number is 10"
    NODES_NUM=$NODES_MAX
fi

#log "$NODES_NUM"

# if already running, do nothing
RUNNING_NODES_NUM=`ps -a | grep "[r]pcaddr" | awk 'END{ print NR }'`
#log "$RUNNING_NODES_NUM"
if [[ "$RUNNING_NODES_NUM" -gt 0 ]]; then
    if [[ "$1" == "$STOP" ]];then
        PID=`ps -a | grep "[r]pcaddr" | awk -v line="$NODES_NUM" 'FNR == line { print $1 }'`

        if [[ "$NODES_NUM" -le "$RUNNING_NODES_NUM" ]];then
            log "Stop node $NODES_NUM, pid = $PID"
            kill -9 $PID
        fi
        exit 0
    fi


    if [[ "$1" == "$APPEND" ]];then
        append_node $RUNNING_NODES_NUM $NODES_NUM
        exit 0
    else
        log "$RUNNING_NODES_NUM nodes are running, kill them before starting new nodes."
        exit 0
    fi
fi

if [[ "$1" == "$NEWACCOUNT" ]];then
    new_account
    exit 0
fi


if [[ "$1" == "$CLEANDB" ]];then
    confirm_do "Are you sure to delete chain database except keystore?[y/N] " clean_db
    exit 0
fi

# execute the command
if [[ "$1" == "$INIT" ]];then
    init_node
else
    append_node 0 $NODES_NUM
fi

exit 0