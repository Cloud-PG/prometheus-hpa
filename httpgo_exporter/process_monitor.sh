#!/bin/bash
# Author: Valentin Kuznetsov
# process_monitor.sh script starts new process_exporter for given
# pattern and prefix. It falls into infinitive loop with given interval
# and restart process_exporter for our pattern process.

usage="Usage: process_monitor.sh <pattern> <prefix> <address> <interval>"
if [ $# -ne 4 ]; then
    echo $usage
    echo ""
    echo "For example, start UNIX process monitoring on address :18883 with 15 sec interval"
    echo ""
    echo "# monitor process with given scitoken pattern"
    echo "process_monitor.sh \".*scitoken\" scitoken \":18883\" 15"
    echo ""
    echo "# monitor crabserver via provide pid file"
    echo "process_monitor.sh /data/srv/state/crabserver/pid scitoken \":18883\" 15"
    exit 1
fi
if [ "$1" == "-h" ] || [ "$1" == "-help" ] || [ "$1" == "--help" ]; then
    echo $usage
    exit 1
fi

# setup our input parameters
pat=$1
prefix=$2
address=$3
interval=$4
while :
do
    # start with empty pid
    pid=""
    # if pat is an existing file, e.g. /data/srv/state/reqmgr2/pid
    # we'll use it to find out group pid
    if [ -f "$pat" ]; then
        gid=`cat $pat`
        pid=`ps -g $gid | grep -v PID | awk '{print $1}' | sort | tail -1`
        if [ -z "$pid" ]; then
            pid=`ps -p $gid | grep -v PID | awk '{print $1}' | sort | tail -1`
        fi
    else
        # find pid of our pattern
        pid=`ps auxw | grep "$pat" | grep -v grep | grep -v process_exporter | grep -v process_monitor | grep -v rotatelogs | tail -1 | awk '{print $2}'`
    fi
    if [ -z "$pid" ]; then
        echo "No pattern '$pat' found"
        sleep $interval
        continue
    fi

    # check if process_exporter is running
    out=`ps auxw | grep "process_exporter -pid $pid -prefix $prefix" | grep -v grep`
    if [ -n "$out" ]; then
        echo "Found existing process_exporter: $out"
        sleep $interval
        continue
    fi

    # find if there is existing process_exporter process
    out=`ps auxw | grep "process_exporter -pid [0-9]* -prefix $prefix" | grep -v grep`
    if [ -n "$out" ]; then
        echo "Killing existing process_exporter: $out"
        prevpid=`echo "$out" | awk '{print $2}'`
        kill -9 $prevpid
    fi

    # start new process_exporter process
    echo "Starting: process_exporter -pid=$pid -prefix $prefix"
    #nohup process_exporter -pid $pid -prefix $prefix -address $address 2>&1 1>& /dev/null < /dev/null &
    process_exporter -pid $pid -prefix $prefix -address $address &

    # sleep our interval for next iteration
    sleep $interval
done
