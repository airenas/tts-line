#!/bin/sh
term_handler() {
  echo Terminating
  if [ $pid -ne 0 ]; then
    kill -SIGTERM "$pid"
    wait "$pid"
    echo Terminated
  fi
  exit 143; # 128 + 15 -- SIGTERM
}

echo Starting
echo Running initial decryption task
./check-decrypt-file -f acronyms.txt 
echo Run main task
./acronyms &
pid="$!"

trap 'kill ${!}; term_handler' SIGTERM

while true
do
  tail -f /dev/null & wait ${!}
done