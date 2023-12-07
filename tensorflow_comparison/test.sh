#!/bin/bash
set -e

./tensorflow_predict.py \
	seq1.fa \
	seq2.fa \
	pairing.txt \
	3
