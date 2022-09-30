#!/bin/bash

OLDREF="${1}"
OUTDIR="${2}"
REF=${OUTDIR}/reference.fa

# OLDREF=/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/refs/combos/axw/axw.fa.gz
# OUTDIR=/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic/out/axw_1/bwa/
# REF=${OUTDIR}/reference.fa
# INDIR=/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic/out/axw_1/trimmomatic/1/


mkdir -p $OUTDIR
rsync $OLDREF $REF
bwa index $REF
