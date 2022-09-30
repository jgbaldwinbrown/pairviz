#!/bin/bash

in1=$1
in2=$2
OUTDIR=$3
outpre=${OUTDIR}/`basename $in1 _R1_001.fastq.gz`

FDATAPATH=$in1
RDATAPATH=$in2
FDATAPATH_TRIMMED=${outpre}_ftrimmed.fq
RDATAPATH_TRIMMED=${outpre}_rtrimmed.fq
FDATAPATH_TRIMMED_UNPAIRED=${outpre}_ftrimmed_unpaired.fq
RDATAPATH_TRIMMED_UNPAIRED=${outpre}_rtrimmed_unpaired.fq

trimmomatic PE \
    -threads 32 \
    -phred33 \
    $FDATAPATH \
    $RDATAPATH \
    $FDATAPATH_TRIMMED \
    $FDATAPATH_TRIMMED_UNPAIRED \
    $RDATAPATH_TRIMMED \
    $RDATAPATH_TRIMMED_UNPAIRED \
    ILLUMINACLIP:/data1/jbrown/human_transmission_distortion_project/scripts/illumina/trimmomatic/adapters.fa:2:30:10 \
    LEADING:20 \
    TRAILING:20 \
    MINLEN:30
    

#java -jar /absolute/path/to/trimmomatic/trimmomatic-0.36.jar PE -threads $CORES -phred33 $FDATAPATH $RDATAPATH $FDATAPATH_TRIMMED $FDATAPATH_TRIMMED_UNPAIRED $RDATAPATH_TRIMMED $RDATAPATH_TRIMMED_UNPAIRED ILLUMINACLIP:/path/to/adapter/sequences.fa:2:30:10 LEADING:20 TRAILING:20 MINLEN:30 CROP:85
