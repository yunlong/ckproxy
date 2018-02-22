
#!/bin/bash

> result.log
date

req_num=1000
key_nums="10"
client_nums="10 20 50 100 200 500 800 1000 1200"

ip=127.0.0.1

i=1
    echo "|$i key|||" >> result.log
    echo "|client_num|request per second| Time per request| time per request" >> result.log
    for j in $client_nums
    do
        echo -n $j >> result.log
        ab -k -n$req_num -c$j -p cloudquery13.txt $ip:8088/cloudquery.php > $i.result.log
        grep -E "Time per request|Requests per second" ./$i.result.log|awk '{print $4}' |xargs|sed "s/[ ]/|/g" >> result.log
        sleep 3;
    done
date
