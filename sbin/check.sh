#!/bin/bash

num=`ps -ef |grep /home/s/safe/skylarproxy/sbin/ckproxy|grep -v grep|grep -v check|wc -l`

if [ ${num} -eq 0 ];then
    #/bin/bash /home/s/safe/skylarproxy/sbin/srvctl init
    /bin/bash /home/s/safe/skylarproxy/sbin/srvctl start -f
fi
