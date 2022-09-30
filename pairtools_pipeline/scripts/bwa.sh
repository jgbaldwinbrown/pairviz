#!/bin/bash

OUTDIR="${1}"
R1="${2}"
R2="${3}"
REF=${OUTDIR}/reference.fa

# OLDREF=/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/refs/combos/axw/axw.fa.gz
# OUTDIR=/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic/out/axw_1/bwa/
# REF=${OUTDIR}/reference.fa
# INDIR=/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic/out/axw_1/trimmomatic/1/


out=${OUTDIR}/`basename $R1 _ftrimmed.fq.gz`.bam
bwa mem -t 32 $REF <(gunzip -c $R1) <(gunzip -c $R2) | samtools view -S -b | samtools sort > $out
