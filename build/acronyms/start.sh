#!/bin/sh
set -e

echo Running startup script

./check-decrypt-file -f acronyms.txt 

exec ./acronyms