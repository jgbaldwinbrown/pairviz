OLDREF="${1}"
INDIR="${2}"
OUTDIR="${3}"
NAMEGLOB="${4}"
REF=${OUTDIR}/reference.fa

# OLDREF=/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/refs/combos/axw/axw.fa.gz
# OUTDIR=/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic/out/axw_1/bwa/
# REF=${OUTDIR}/reference.fa
# INDIR=/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic/out/axw_1/trimmomatic/1/


mkdir -p $OUTDIR
rsync $OLDREF $REF
bwa index $REF
find $INDIR -name ${NAMEGLOB} | \
sort | \
paste - - | \
while read i
do
    a=`echo $i | cut -d ' ' -f 1`
    b=`echo $i | cut -d ' ' -f 2`
    out=${OUTDIR}/`basename $a _ftrimmed.fq.gz`.bam
    bwa mem -t 32 $REF <(gunzip -c $a) <(gunzip -c $b) | samtools view -S -b | samtools sort > $out
    samtools index $out
done

