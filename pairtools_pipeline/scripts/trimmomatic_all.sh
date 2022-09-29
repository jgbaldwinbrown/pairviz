OUTDIR="${1}"
INDIR="${2}"
NAMEGLOB="${3}"
# OUTDIR=/media/jgbaldwinbrown/jim_work1/jgbaldwinbrown/Documents/work_stuff/drosophila/homologous_hybrid_mispairing/hic/from_lonepeak/mini2/axw_1/trimmomatic/1/
# INDIR=/media/jgbaldwinbrown/jim_work1/jgbaldwinbrown/Documents/work_stuff/drosophila/homologous_hybrid_mispairing/hic/from_lonepeak/mini2/axw_1/fq/

mkdir -p $OUTDIR
# find $INDIR -name '*.fastq.gz' | \
find $INDIR -name ${NAMEGLOB} | \
sort | \
paste - - | \
while read i
do
    a=`echo $i | cut -d ' ' -f 1`
    b=`echo $i | cut -d ' ' -f 2`
    out=${OUTDIR}/`basename $a _R1_001.fastq.gz`
    bash scripts/hic/trimmomatic.sh <(gunzip -c $a) <(gunzip -c $b) $out
    ls -d ${out}* | grep -vE 'gz$' | xargs pigz -p 32
    #echo $i $a $b $out
done
