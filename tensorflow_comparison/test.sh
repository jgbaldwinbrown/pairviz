#!/bin/bash
set -e

./check.sh

./tensorflow_predict.py \
	seq1.fa \
	seq2.fa \
	pairing.txt \
	3 \
	25
