#!/bin/bash
set -e

(cd go_pairviz && go build)
# cat mini.pairs | ./go_pairviz/go_pairviz -r regions.txt
cat mini.pairs | ./go_pairviz/go_pairviz -c
cat mini.pairs | ./pairviz.py -c -g 200000000

cat mini.pairs | ./go_pairviz/go_pairviz -w 1000000 -s 100000
cat mini.pairs | ./pairviz.py -w 1000000 -s 100000 -g 200000000
