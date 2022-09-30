#!/bin/bash

set -e

IN=$1
SIZE=$2
OUT=$3

# source /data1/jbrown/local_programs/pairtools/anaconda_on.sh
pairtools parse -c $SIZE -o $OUT --drop-sam $IN
