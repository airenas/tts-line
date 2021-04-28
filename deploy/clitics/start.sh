#!/bin/sh
set -e

echo Running startup script

./check-decrypt-file -f phrases.txt 

exec ./clitics