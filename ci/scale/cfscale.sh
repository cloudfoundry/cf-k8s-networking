#!/bin/bash

udate ()
{
  date +%s
}

user()
{
  echo "$(udate),$1,STARTED,"
  TARGET_URL=bin-scaletest-$1.$2

  cf map-route bin-$1 $2 --hostname bin-scaletest-$1

  echo "$(udate),$1,POLLING,"
  until [ $(curl -s -o /dev/null -w "%{http_code}" http://$TARGET_URL/status/200) -eq 200 ]; do true; done

  echo "$(udate),$1,SUCCESS,"

  lastfail=$(udate)
  for ((i=120; i>0; i--)); do
    sleep 1 &
    status=$(curl -s -o /dev/null -w "%{http_code}" http://$TARGET_URL/status/200)
    if [ $status -ne 200 ]; then
      lastfail=$(udate)
      echo "$(udate),$1,FAILURE,$status"
    fi
    wait
  done

  echo "$lastfail,$1,COMPLETED,"
}

userfactory ()
{
  for ((count = 0; count < $1; count++)); do
    user $count $2 &
    sleep 10
  done

  wait
  echo "User factory complete"
}

echo "Cleaning up routes just in case..."
domain="$(cf domains | tail -n1 | awk '{print $1}')"
cf routes | grep scaletest | awk '{print $2}' | xargs -I{} -n1 cf delete-route $domain --hostname {} -f

userfactory $1 $domain | tee scale-$1.log

echo "stamp,usernum,event,status" > scale-data.csv
cat scale-$1.log | grep "^\d\d\d\d\d" >> scale-data.csv

sqlite3 scale.db -cmd \
  "drop table t" \
  ".mode csv" \
  ".import scale-data.csv t" \
  "select b.stamp - a.stamp tim from t a join t b on a.usernum = b.usernum and a.event='POLLING' and b.event='COMPLETED' order by tim asc;"

echo "You just saw all the latencies. If you did 100, the p95 is 5th from the bottom."
