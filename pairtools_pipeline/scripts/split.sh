#!/bin/bash

NCORES="$1"
IN="$2"
OUTPRE="$3"
SPLITLINES="$4"

pigz -d -p "${NCORES}" -c <${IN} | \
split --lines="${SPLITLINES}" --additional-suffix=.fastq - "${OUTPRE}"
pigz -p "${NCORES}" *.fastq
