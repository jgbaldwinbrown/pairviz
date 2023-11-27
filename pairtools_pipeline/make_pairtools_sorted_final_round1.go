package main

import (
	"flag"
	"path/filepath"
	"time"
	"bufio"
	"compress/gzip"
	"regexp"
	"github.com/jgbaldwinbrown/makem"
	"fmt"
	"os"
)

type Params struct {
	Name string
	Nsplits int64
	LinesPerSplit int64
	Ref string
	Indir string
	Refdir string
	Outdir string
	Scriptdir string
}

// 125530532

func AddSplit(mf *makem.MakeData, name string, nsplits, linesPerSplit int64, indir, outdir, scriptdir string) {
	var r makem.Recipe
	target := filepath.Clean(outdir + "/fq_split/split.done")
	deps := []string {
		filepath.Clean(indir + "/" + name + "_R1_001.fastq.gz"),
		filepath.Clean(indir + "/" + name + "_R2_001.fastq.gz"),
	}
	r.AddTargets(target)
	r.AddDeps(deps...)
	r.AddScripts(
		"mkdir -p `dirname $@`",
		filepath.Clean(scriptdir + "/split.sh") + " 32 " + deps[0] + " " + outdir + "/fq_split/" + name + "_R1_001_split_ " + fmt.Sprintf("%v", linesPerSplit),
		filepath.Clean(scriptdir + "/split.sh") + " 32 " + deps[1] + " " + outdir + "/fq_split/" + name + "_R2_001_split_ " + fmt.Sprintf("%v", linesPerSplit),
		"touch $@",
	)
	mf.Add(r)
}

// outpre=${OUTDIR}/`basename $in1 _R1_001.fastq.gz`

func AddTrim(mf *makem.MakeData, name string, nsplits int64, outdir, scriptdir string) {

	// axw_ftrimmed_split_0008.fq.gz.bam
	splitdir := filepath.Clean(outdir + "/fq_split/") + "/"
	trimdir := filepath.Clean(outdir + "/trimmomatic") + "/"

	var i int64
	for i=0; i<nsplits; i++ {
		var r makem.Recipe
		target := filepath.Clean(fmt.Sprintf("%v/trimmomatic_%04d_done.txt", trimdir, i))
		r.AddTargets(target)
		dep := filepath.Clean(splitdir + "/split.done")
		r.AddDeps(dep)
		// nameglob := CleanSlash(fmt.Sprintf("%v_R?_001_split_%04d.fastq.gz", name, i))
		r1 := filepath.Clean(fmt.Sprintf("%v/%v_R1_001_split_%04d.fastq.gz", splitdir, name, i))
		r2 := filepath.Clean(fmt.Sprintf("%v/%v_R2_001_split_%04d.fastq.gz", splitdir, name, i))
		r.AddScripts(
			"mkdir -p `dirname $@`",
			filepath.Clean(scriptdir + "/trimmomatic_one.sh") + " " + r1 + " " + r2 + " " + trimdir,
			"touch $@",
		)
		mf.Add(r)
	}
}

func AddBwaRef(mf *makem.MakeData, name, ref, refdir, outdir, scriptdir, run string) {
	oldref := filepath.Clean(refdir + "/" + ref + "_ecoli.fa.gz")
	bwasplitdir := filepath.Clean(outdir + "/bwa_split") + "/"

	if run == "hic4" {
		oldref = filepath.Clean(refdir + "/" + ref + ".fa.gz")
	}

	var r makem.Recipe

	r.AddTargets(filepath.Clean(bwasplitdir + "/reference.fa.done"))
	r.AddDeps(oldref)
	r.AddScripts(
		"mkdir -p `dirname $@`",
		filepath.Clean(scriptdir + "/bwaref.sh") + " " + oldref + " " + bwasplitdir,
		"touch $@",
	)

	mf.Add(r)
}

func AddBwa(mf *makem.MakeData, name string, nsplits int64, ref, outdir, scriptdir string) {
	trimdir := filepath.Clean(outdir + "/trimmomatic") + "/"
	bwasplitdir := filepath.Clean(outdir + "/bwa_split") + "/"

	var i int64
	for i=0; i<nsplits; i++ {
		var r makem.Recipe
		target := filepath.Clean(fmt.Sprintf("%v/%v_R1_001_split_%04d.bam.done", bwasplitdir, name, i))
		r.AddTargets(target)
		dep := filepath.Clean(fmt.Sprintf("%v/trimmomatic_%04d_done.txt", trimdir, i))
		refdep := filepath.Clean(bwasplitdir + "/reference.fa.done")
		r.AddDeps(dep, refdep)

		r1 := filepath.Clean(fmt.Sprintf("%v/%v_R1_001_split_%04d.fastq.gz_ftrimmed.fq.gz", trimdir, name, i))
		r2 := filepath.Clean(fmt.Sprintf("%v/%v_R1_001_split_%04d.fastq.gz_rtrimmed.fq.gz", trimdir, name, i))
		r.AddScripts(
			"mkdir -p `dirname $@`",
			filepath.Clean(scriptdir + "/bwa.sh") + " " + bwasplitdir + " " + r1 + " " + r2,
			"touch $@",
		)
		mf.Add(r)
	}
}

func AddMerge(mf *makem.MakeData, name string, nsplits int64, outdir string) {
	var r makem.Recipe
	bwadir := filepath.Clean(outdir + "/bwa") + "/"
	bwasplitdir := filepath.Clean(outdir + "/bwa_split") + "/"
	output := filepath.Clean(fmt.Sprintf("%v/full.bam", bwadir))
	target := filepath.Clean(fmt.Sprintf("%v/full.bam.done", bwadir))
	inglob := filepath.Clean(bwasplitdir + "/*.bam")
	r.AddTargets(target)
	var deps []string

	var i int64
	for i=0; i<nsplits; i++ {
		deps = append(deps, filepath.Clean(fmt.Sprintf("%v/%v_R1_001_split_%04d.bam.done", bwasplitdir, name, i)))
	}
	r.AddDeps(deps...)
	r.AddScripts(
		"mkdir -p `dirname $@`",
		filepath.Clean("samtools merge - " + inglob + " | samtools sort -n > " + output),
		"touch $@",
	)
	mf.Add(r)
}

func AddFullBai(mf *makem.MakeData, name, outdir string) {
	var r makem.Recipe

	bwadir := filepath.Clean(outdir + "/bwa") + "/"
	bam := filepath.Clean(bwadir + "/full.bam")
	r.AddTargets(bam + ".bai.done")
	r.AddDeps(bam + ".done")
	r.AddScripts(
		"mkdir -p `dirname $@`",
		"samtools index " + bam,
		"touch $@",
	)

	mf.Add(r)
}

func AddPairtools(mf *makem.MakeData, name, refdir, ref, outdir, scriptdir string) {
	var r makem.Recipe

	bwadir := filepath.Clean(outdir + "/bwa" + "/")
	pairdir := filepath.Clean(outdir + "/pairtools" + "/")
	bam := filepath.Clean(bwadir + "/full.bam")
	r.AddTargets(filepath.Join(pairdir, "pairtools_done.txt"))
	r.AddDeps(bam + ".done")
	pt := filepath.Clean(scriptdir + "/pairtools1.sh")

	chrlens := filepath.Clean(refdir + "/" + ref + "_ecoli.fa.gz.chrlens.txt")
	output := filepath.Clean(pairdir + "/" + name + "_to_" + ref + "ref.pairs")
	r.AddScripts(
		"mkdir -p `dirname $@`",
		pt + " " + bam + " " + chrlens + " " + output,
		"touch $@",
	)

	mf.Add(r)

}

func UpdateNsplits(p *Params) error {
	if p.Nsplits < 1 {
		var err error
		p.Nsplits, err = CalcSplitsFromFq(*p)
		if err != nil {
			return err
		}
	}
	return nil
}

func AddRun(mf *makem.MakeData, p Params, run string) error {
	if err := UpdateNsplits(&p); err != nil {
		return err
	}

	AddSplit(mf, p.Name, p.Nsplits, p.LinesPerSplit, p.Indir, p.Outdir, p.Scriptdir)
	AddTrim(mf, p.Name, p.Nsplits, p.Outdir, p.Scriptdir)
	AddBwaRef(mf, p.Name, p.Ref, p.Refdir, p.Outdir, p.Scriptdir, run)
	AddBwa(mf, p.Name, p.Nsplits, p.Ref, p.Outdir, p.Scriptdir)
	AddMerge(mf, p.Name, p.Nsplits, p.Outdir)
	// AddFullBai(mf, p.Name, p.Outdir)
	AddPairtools(mf, p.Name, p.Refdir, p.Ref, p.Outdir, p.Scriptdir)

	return nil
}

func MakeMakefile(params []Params, run string) makem.MakeData {
	var mf makem.MakeData
	for _, p := range params {
		AddRun(&mf, p, run)
	}
	return mf
}

func MakeSplitPath(p Params, i int64) string {
	return fmt.Sprintf("runscripts/run_%v_split_%04d.sh", p.Name, i)
}

func MakeSplit(p Params, i int64) error {
	if err := UpdateNsplits(&p); err != nil {
		return err
	}
	os.Mkdir("runscripts", 0777)
	spath := MakeSplitPath(p, i)
	w, err := os.Create(spath)
	if err != nil {
		return err
	}
	defer w.Close()

	target := filepath.Clean(fmt.Sprintf("%v/%v_R1_001_split_%04d.bam.done", p.Outdir + "/bwa_split/", p.Name, i))

	fmt.Fprintf(
		w,
		`#!/bin/bash
set -e
#SBATCH -t 72:00:00    #max:    72 hours (24 on ash)
#SBATCH -N 1          #format: count or min-max
#SBATCH -A owner-guest    #values: yandell, yandell-em (ember), ucgd-kp (kingspeak)
#SBATCH -p notchpeak-guest    #kingspeak, ucgd-kp, kingspeak-freecycle, kingspeak-guest
#SBATCH -J %v_split_%v        #Job name

NUM_CORES="${SLURM_CPUS_ON_NODE}"

module load trimmomatic/0.39
module load bwa/2020_03_19
module load samtools/1.12
module load python/3.10.3

make -j $NUM_CORES %v
`,
		p.Name,
		i,
		target,
	)
	return nil
}

func MakeSplits(p Params) error {
	spath := fmt.Sprintf("runscripts/run_%v_splits.sh", p.Name)
	w, err := os.Create(spath)
	if err != nil {
		return err
	}
	defer w.Close()

	var i int64
	for i=0; i<p.Nsplits; i++ {
		splitpath := MakeSplitPath(p, i)
		err = MakeSplit(p, i)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "bash %v\n", splitpath)
	}
	return nil
}

func MakeSplitsCluster(p Params) error {
	spath := fmt.Sprintf("runscripts/run_%v_splits_cluster.sh", p.Name)
	w, err := os.Create(spath)
	if err != nil {
		return err
	}
	defer w.Close()

	var i int64
	for i=0; i<p.Nsplits; i++ {
		splitpath := MakeSplitPath(p, i)
		err = MakeSplit(p, i)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "sbatch %v\n", splitpath)
	}
	return nil
}

func MakeEnds(p Params) error {
	spath := fmt.Sprintf("runscripts/run_%v_end.sh", p.Name)
	w, err := os.Create(spath)
	if err != nil {
		return err
	}
	defer w.Close()

	target := filepath.Clean(p.Outdir + "/pairtools/pairtools_done.txt")

	fmt.Fprintf(
		w,
		`#!/bin/bash
set -e
#SBATCH -t 72:00:00    #max:    72 hours (24 on ash)
#SBATCH -N 1          #format: count or min-max
#SBATCH -A owner-guest    #values: yandell, yandell-em (ember), ucgd-kp (kingspeak)
#SBATCH -p notchpeak-guest    #kingspeak, ucgd-kp, kingspeak-freecycle, kingspeak-guest
#SBATCH -J %v_end        #Job name

NUM_CORES="${SLURM_CPUS_ON_NODE}"

module load trimmomatic/0.39
module load bwa/2020_03_19
module load samtools/1.12
module load python/3.10.3

make -j $NUM_CORES %v
`,
		p.Name,
		target,
	)
	return nil
}

func MakeStarts(p Params) error {
	spath := fmt.Sprintf("runscripts/run_%v_start.sh", p.Name)

	w, err := os.Create(spath)
	if err != nil {
		return err
	}
	defer w.Close()

	target := filepath.Clean(p.Outdir + "/bwa_split/reference.fa.done")
	target2 := filepath.Clean(p.Outdir + "/fq_split/split.done")

	fmt.Fprintf(
		w,
		`#!/bin/bash
set -e
#SBATCH -t 72:00:00    #max:    72 hours (24 on ash)
#SBATCH -N 1          #format: count or min-max
#SBATCH -A owner-guest    #values: yandell, yandell-em (ember), ucgd-kp (kingspeak)
#SBATCH -p notchpeak-guest    #kingspeak, ucgd-kp, kingspeak-freecycle, kingspeak-guest
#SBATCH -J %v_start        #Job name

NUM_CORES="${SLURM_CPUS_ON_NODE}"

module load trimmomatic/0.39
module load bwa/2020_03_19
module load samtools/1.12
module load python/3.10.3

make -j $NUM_CORES %v %v
`,
		p.Name,
		target,
		target2,
	)
	return nil
}

func BuildParams(names []string, indirPrefix, refdirPrefix, outdirPrefix, scriptdir string) []Params {
	re := regexp.MustCompile(`_(adult|sal|brain|fat)`)

	var ps []Params
	for _, name := range names {
		fmt.Println("original:", name, "replaced:", re.ReplaceAllString(name, ""))
		ref := re.ReplaceAllString(name, "")
		ps = append(ps, Params {
			Name: name,
			Nsplits: -1,
			LinesPerSplit: 100000000,
			Ref: ref,
			Indir: indirPrefix + "/" + name + "_full/",
			Refdir: refdirPrefix + "/" + ref + "/",
			Outdir: outdirPrefix + "/" + name + "_1/",
			Scriptdir: scriptdir,
		})
	}
	return ps
}

func VerySmallParams(run string) []Params {
	names := []string { "s14xw501_sal", "s14xw501_adult" }

	indirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/data/21326R/"
	refdirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/refs/combos/"
	outdirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic5_final_ecoli/out/"
	scriptdir := "scripts/"

	if run == "hic4" {
		outdirPrefix = "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/out/"
	}

	return BuildParams(names, indirPrefix, refdirPrefix, outdirPrefix, scriptdir)
}

func SmallParams(run string) []Params {
	names := []string { "s14xw501_sal", "s14xw501_adult", "iso1xsaw_sal", "sawxw501_sal"  }

	indirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/data/21326R/"
	refdirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/refs/combos/"
	outdirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic5_final_ecoli/out/"
	scriptdir := "scripts/"

	if run == "hic4" {
		outdirPrefix = "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/out/"
	}

	return BuildParams(names, indirPrefix, refdirPrefix, outdirPrefix, scriptdir)
}

func MidParams(run string) []Params {
	names := []string {
		"s14xw501_sal", "s14xw501_adult", "iso1xsaw_sal", "sawxw501_sal",
		"iso1xa4_adult", "iso1xa4_brain", "iso1xa4_fat", "iso1xa4_sal",
		"iso1xa7_sal", "iso1xa7_adult", "m252xw501_adult", "m252xw501_sal",
	}

	indirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/data/21326R/"
	refdirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/refs/combos/"
	outdirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic5_final_ecoli/out/"
	scriptdir := "scripts/"

	if run == "hic4" {
		outdirPrefix = "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/out/"
	}

	return BuildParams(names, indirPrefix, refdirPrefix, outdirPrefix, scriptdir)
}

func FullParams(run string) []Params {
	names := []string {
		"s14xw501_sal", "s14xw501_adult", "iso1xsaw_sal", "sawxw501_sal",
		"iso1xa4_adult", "iso1xa4_brain", "iso1xa4_fat", "iso1xa4_sal",
		"iso1xa7_sal", "iso1xa7_adult", "m252xw501_adult", "m252xw501_sal",
		"iso1xlhr_sal", "hmrxw501_sal",
		"a7xs14_sal", "a7xs14_adult",
		"iso1xw501_sal", "iso1xw501_adult", "iso1xw501_brain", "iso1xw501_fat",
	}

	indirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/data/21326R/"
	refdirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/refs/combos/"
	outdirPrefix := "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic5_final_ecoli/out/"
	scriptdir := "scripts/"

	if run == "hic4" {
		outdirPrefix = "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/out/"
	}

	return BuildParams(names, indirPrefix, refdirPrefix, outdirPrefix, scriptdir)
}

type GzScanner struct {
	fp *os.File
	gr *gzip.Reader
	*bufio.Scanner
}

func (s *GzScanner) Close() error {
	var err error
	if e := s.gr.Close(); e != nil {
		if err == nil {
			err = e
		}
	}
	if e := s.fp.Close(); e != nil {
		if err == nil {
			err = e
		}
	}
	return err
}

func OpenGzScanner(gzpath string) (*GzScanner, error) {
	fp, err := os.Open(gzpath)
	if err != nil {
		return nil, err
	}

	gr, err := gzip.NewReader(fp)
	if err != nil {
		fp.Close()
		return nil, err
	}

	s := bufio.NewScanner(gr)
	s.Buffer([]byte{}, 1e12)

	return &GzScanner{fp, gr, s}, nil
}

func AlwaysCountGzLines(gzpath string) (int64, error) {
	s, err := OpenGzScanner(gzpath)
	if err != nil {
		return 0, err
	}

	var i int64 = 0
	for s.Scan() {
		if s.Err() != nil {
			return 0, s.Err()
		}
		i++
	}
	return i, nil
}

func ReadCount(cpath string) (int64, error) {
	r, err := os.Open(cpath)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	var count int64
	_, err = fmt.Fscanf(r, "%v", &count)
	return count, err
}

func WriteCount(count int64, opath string) error {
	if err := os.MkdirAll(filepath.Dir(opath), 0777); err != nil {
		return err
	}

	w, err := os.Create(opath)
	if err != nil {
		return err
	}
	defer w.Close()
	_, err = fmt.Fprintln(w, count)
	return err
}

func Touch(opath string) error {
	_, err := os.Stat(opath)
	if os.IsNotExist(err) {
		file, err := os.Create(opath)
		if err != nil {
			return err
		}
		file.Close()
		return nil
	}

	now := time.Now().Local()
	err = os.Chtimes(opath, now, now)
	if err != nil {
		return err
	}
	return nil
}

func CountLines(fapath, outpre string, write bool) (int64, error) {
	countpath := outpre + "_count.txt"
	countedpath := outpre + "_count.done"

	var count int64

	_, err := os.Stat(countedpath)
	if os.IsNotExist(err) {
		count, err = AlwaysCountGzLines(fapath)
	} else {
		count, err = ReadCount(countpath)
	}
	if err != nil {
		return 0, err
	}

	if write {
		err = WriteCount(count, countpath)
		if err != nil {
			return 0, err
		}
		err = Touch(countedpath)
		if err != nil {
			return 0, err
		}
	}

	return count, nil
}

func CalcSplits(linesPerSplit, lines int64) int64 {
	splits := lines / linesPerSplit
	if (lines % linesPerSplit != 0) {
		splits++
	}
	return splits
}

func CalcSplitsFromFq(p Params) (nsplits int64, err error) {
	outpre := filepath.Clean(p.Outdir + "/" + p.Name)
	fapath := filepath.Clean(p.Indir + "/" + p.Name + "_R1_001.fastq.gz")

	fmt.Fprintf(os.Stderr, "started counting lines from %v, putting result in %v\n", fapath, outpre)
	lines, err := CountLines(fapath, outpre, true)
	if err != nil {
		return 0, err
	}

	return CalcSplits(p.LinesPerSplit, lines), nil
}

func main() {
	run := flag.String("r", "", "run (try using \"hic4\")")
	flag.Parse()
	params := FullParams(*run)
	for i, _ := range params {
		if err := UpdateNsplits(&params[i]); err != nil {
			panic(err)
		}
	}

	mf := MakeMakefile(params, *run)
	mfFile, err := os.Create("Makefile")
	if err != nil {
		panic(err)
	}
	defer mfFile.Close()
	mf.Fprint(mfFile)
	os.Mkdir("runscripts", 0777)
	for _, param := range params {
		err = MakeStarts(param)
		if err != nil {
			panic(err)
		}
		err := MakeSplits(param)
		if err != nil {
			panic(err)
		}
		err = MakeSplitsCluster(param)
		if err != nil {
			panic(err)
		}
		err = MakeEnds(param)
		if err != nil {
			panic(err)
		}
	}
}
