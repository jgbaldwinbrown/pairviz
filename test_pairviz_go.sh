#!/bin/bash
set -e

(cd go_pairviz && go build)
cat mini.pairs | ./go_pairviz/go_pairviz -r regions.txt
cat mini.pairs | ./go_pairviz/go_pairviz -c
