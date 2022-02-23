#!/bin/bash
set -e

(cd go_pairviz && go build)
gunzip -c testdata2.pairs.gz > testdata2.pairs

time (<testdata2.pairs ./go_pairviz/go_pairviz -w 1000000 -s 100000 > /dev/null)
time (<testdata2.pairs ./pairviz.py -w 1000000 -s 100000 -g 200000000 > /dev/null)

<testdata2.pairs ./go_pairviz/go_pairviz -w 1000000 -s 100000 > gotestout.txt
<testdata2.pairs ./pairviz.py -w 1000000 -s 100000 -g 200000000 > pytestout.txt

<testdata2.pairs ./go_pairviz/go_pairviz -r <(echo "3R	0	1000000") > goregiontestout.txt

rm testdata2.pairs
