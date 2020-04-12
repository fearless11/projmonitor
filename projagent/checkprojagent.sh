#!/bin/sh
## desc: monitor the cpu and memory of the projagent.

source /etc/profile

#check %cpu、RES of projagent 
PID=`pgrep projagent`
DIR="/data/to8to/tools/projagent"
mkdir -p $DIR/log

if [[ ${PID} == "" ]];then
   echo "[`date +%F-%T`] projagent start" &>> $DIR/log/projagent.log
   cd $DIR
   ./control start &
   exit
fi


CPU=`top -p $PID -b -d 1 -n 1 |grep %CPU -A 1 |grep -v %CPU | awk '{ print  $9}'`
RES=`top -p $PID -b -d 1 -n 1 |grep %CPU -A 1 |grep -v %CPU | awk '{ print  $6}'|grep m |awk -Fm '{print $1}'`


if [[ $(echo "$CPU > 100.0"|bc) == 1 ]] || [[ $RES > 80 ]];then
        sleep 10
        CPU=`top -p $PID -b -d 1 -n 1 |grep %CPU -A 1 |grep -v %CPU | awk '{ print  $9}'`
        RES=`top -p $PID -b -d 1 -n 1 |grep %CPU -A 1 |grep -v %CPU | awk '{ print  $6}'|grep m |awk -Fm '{print $1}'`
        if [[ $CPU > 100.0 ]] || [[ $RES > 80 ]];then
          echo "[`date +%F-%T`] $CPU $RES projagent restart" &>> $DIR/log/projagent.log
              cd $DIR
              ./control restart &
        fi
fi

NUM=`pgrep projagent |wc -l`
if [[ $NUM > 1 ]]; then killall projagent; killall jstat; fi


LOGSIZE=`du -sh /data/to8to/tools/projagent/log/app.log  |grep M |awk -F'M' '{print $1}'`
if [[ $LOGSIZE > 500 ]];then
     echo > /data/to8to/tools/projagent/log/app.log
fi

jstatnum=`pgrep -f jstat|wc -l`
if [[ $jstatnum -gt 10 ]]; then killall jstat; fi
