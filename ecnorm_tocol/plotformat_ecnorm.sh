#!/bin/bash
set -e

cat | \
jq '"\(.Chr)_\(.Genome)\t\(.Start)\t\(.End)\t\(.TargetProp)"' | \
sed 's/"//g' | sed 's/\\t/\t/g'
