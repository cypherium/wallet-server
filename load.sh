#!/bin/bash
echo "=="$(date  +"%Y-%m-%d %H:%M:%S")"=="
mkdir log  >/dev/null 2>&1

CMD='./load.sh'
STOPEDFILE=.stoped.flag
appname='scan'

start() {
    echo "--------------------";
    echo "Starting private network";

    nohup ./bin/$appname >> ./coutput.log &
    sleep 0.1;
    echo "-----";
    ps -elf | grep -E "./bin/$appname$" | grep -v grep;
    echo "DONE";
    ps|grep $appname
    rm -f $STOPEDFILE
}

stop() {
    echo "Stop $appname application";
    rm log/*
    pkill -9 $appname
    killflag=0
    pidlist=`ps -elf | grep -E "./bin/$appname" | grep -v grep | awk '{print $4}'`;
    for pid in $pidlist;
    do
        killflag=1
        kill $pid ;
    done;

    if [ $killflag -eq 1 ]; then
        sleep 0.1;
        echo "-----";
        ps -elf | grep $appname| grep -v grep;
        echo "stop done"
    else
        echo './bin/$appname is not running';
    fi
    touch $STOPEDFILE
}

restart() {
    stop;
    start;
}

install() {
    $appnamepath=`pwd`;
    if [ ! -f "$appnamepath/bin/$appname" ]; then
        echo 'error!!!'
        echo 'not in "$appnamepath" directory?'
        echo 'error!!!'
        exit 1
    fi
    cronconf="* * * * * cd $appnamepath && ./load.sh check >> log/cron.log&"
    exist=`crontab -l | grep -F "$cronconf" | grep -v 'grep'`
    if [ "$exist" = "" ]; then
        (crontab -l 2>/dev/null | grep -Fv "$cronconf"; echo "$cronconf") | crontab -
        echo "add crontab[$cronconf] done"
    else
        echo "crontab[$cronconf] have exist! error"
    fi
}

check() {
    if [ -f ./$STOPEDFILE ]; then
        echo "it stoped, do nothing!!"
        exit 0
    fi
    pidlist=`ps -elf | grep -E "./bin/$appname$" | grep -v grep | awk '{print $4}'`
    if [ "$pidlist" = "" ] ;then
        echo "$appname not exist, try treload";
        restart;
    else
        echo "$appname is running, do nothing...";
    fi
}


arg1=$1
if [ ! "$0" = "$CMD" ]; then
    echo "error!! "
    echo "You should in correct directory "
    echo "and run command './load.sh'"
    echo "error!! "
    arg1="error"
fi

case "$arg1" in
start)
    start
    ;;
stop)
    stop
    ;;
install)
    install
    ;;
check)
    check
    ;;
restart|reload)
    restart
    ;;
*)
    echo "Usage: $CMD {install|check|start|stop|restart}"
    exit 1
esac
