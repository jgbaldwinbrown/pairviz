package main

import (
	"github.com/jgbaldwinbrown/makem"
	"fmt"
	"os"
	"regexp"
)

type Params struct {
	Name string
	Nsplits int
	Ref string
	Indir string
	Refdir string
	Outdir string
	Scriptdir string
}

// 125530532

func CleanSlash(s string) string {
	re := regexp.MustCompile(`/+`)
	return re.ReplaceAllString(s, "/")
}

func AddSplit(mf *makem.MakeData, name string, nsplits int, indir, outdir, scriptdir string) {
	var r makem.Recipe
	target := outdir + "/fq_split/split.done"
	deps := []string {
		CleanSlash(indir + "/" + name + "_R1_001.fastq.gz"),
		CleanSlash(indir + "/" + name + "_R2_001.fastq.gz"),
	}
	r.AddTargets(target)
	r.AddDeps(deps...)
	r.AddScripts(
		"mkdir -p `dirname $@`",
		CleanSlash(scriptdir + "/fq_split/split.sh 32 " + deps[0] + " " + outdir + "/" + name + "_R1_001_split_ 125530532"),
		CleanSlash(scriptdir + "/fq_split/split.sh 32 " + deps[1] + " " + outdir + "/" + name + "_R2_001_split_ 125530532"),
		"touch $@",
	)
	mf.Add(r)
}

func AddTrim(mf *makem.MakeData, name string, nsplits int, outdir, scriptdir string) {

	// axw_ftrimmed_split_0008.fq.gz.bam
	splitdir := CleanSlash(outdir + "/fq_split/")
	trimdir := CleanSlash(outdir + "/trimmomatic/")

	for i:=0; i<nsplits; i++ {
		var r makem.Recipe
		target := CleanSlash(fmt.Sprintf("%v/trimmomatic_%04d_done.txt", trimdir, i))
		r.AddTargets(target)
		dep := CleanSlash(splitdir + "/split.done")
		r.AddDeps(dep)
		// nameglob := CleanSlash(fmt.Sprintf("%v_R?_001_split_%04d.fastq.gz", name, i))
		r1 := CleanSlash(fmt.Sprintf("%v_R1_001_split_%04d.fastq.gz", name, i))
		r2 := CleanSlash(fmt.Sprintf("%v_R2_001_split_%04d.fastq.gz", name, i))
		r.AddScripts(
			"mkdir -p `dirname $@`",
			CleanSlash(scriptdir + "/trimmomatic_one.sh " + r1 + " " + r2 + " " + trimdir),
			"touch $@",
		)
		mf.Add(r)
	}
}

func AddBwaRef(mf *makem.MakeData, name, ref, refdir, outdir, scriptdir string) {
	oldref := CleanSlash(refdir + "/" + ref + ".fa.gz")
	bwasplitdir := CleanSlash(outdir + "/bwa_split/")

	var r makem.Recipe

	r.AddTargets(CleanSlash(bwasplitdir + "/reference.fa.done"))
	r.AddDeps(oldref)
	r.AddScripts(
		"mkdir -p `dirname $@`",
		CleanSlash(scriptdir + "/bwaref.sh " + oldref + " " + bwasplitdir),
		"touch $@",
	)

	mf.Add(r)
}

func AddBwa(mf *makem.MakeData, name string, nsplits int, ref, refdir, outdir, scriptdir string) {
	trimdir := CleanSlash(outdir + "/trimmomatic/")
	bwasplitdir := CleanSlash(outdir + "/bwa_split/")

	for i:=0; i<nsplits; i++ {
		var r makem.Recipe
		target := CleanSlash(fmt.Sprintf("%v/%v_R1_001_split_%04d.bam.done", bwasplitdir, name, i))
		r.AddTargets(target)
		dep := CleanSlash(fmt.Sprintf("%v/trimmomatic_%04d_done.txt", trimdir, i))
		refdep := CleanSlash(bwasplitdir + "/reference.fa.done")
		r.AddDeps(dep, refdep)
		// nameglob := CleanSlash(fmt.Sprintf("%v_R?_001_split_%04d_?trimmed.fastq.gz", name, i))
		r1 := CleanSlash(fmt.Sprintf("%v/%v_R1_001_split_%04d_ftrimmed.fastq.gz", trimdir, name, i))
		r2 := CleanSlash(fmt.Sprintf("%v/%v_R1_001_split_%04d_rtrimmed.fastq.gz", trimdir, name, i))
		r.AddScripts(
			"mkdir -p `dirname $@`",
			CleanSlash(scriptdir + "/bwa.sh " + bwasplitdir + " " + r1 + " " + r2),
			"touch $@",
		)
		mf.Add(r)
	}
}

func AddMerge(mf *makem.MakeData, name string, nsplits int, outdir string) {
	var r makem.Recipe
	bwadir := CleanSlash(outdir + "/bwa/")
	bwasplitdir := CleanSlash(outdir + "/bwa_split/")
	output := CleanSlash(fmt.Sprintf("%v/full.bam", bwadir))
	target := CleanSlash(fmt.Sprintf("%v/full.bam.done", bwadir))
	inglob := CleanSlash(bwasplitdir + "/*.bam")
	r.AddTargets(target)
	var deps []string
	for i:=0; i<nsplits; i++ {
		deps = append(deps, CleanSlash(fmt.Sprintf("%v/%v_R1_001_split_%04d.bam.done", bwasplitdir, name, i)))
	}
	r.AddDeps(deps...)
	r.AddScripts(
		"mkdir -p `dirname $@`",
		CleanSlash("samtools merge - " + inglob + " | samtools sort > " + output),
		"touch $@",
	)
	mf.Add(r)
}

func AddFullBai(mf *makem.MakeData, name, outdir string) {
	var r makem.Recipe

	bwadir := CleanSlash(outdir + "/bwa/")
	bam := CleanSlash(bwadir + "/full.bam")
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

	bwadir := CleanSlash(outdir + "/bwa/")
	pairdir := CleanSlash(outdir + "/pairtools/")
	bam := CleanSlash(bwadir + "/full.bam")
	r.AddTargets(pairdir + "pairtools_done.txt")
	r.AddDeps(bam + ".bai.done")
	pt := CleanSlash(scriptdir + "/pairtools1.sh")

	chrlens := CleanSlash(refdir + "/" + name + ".fa.gz.chrlens.txt")
	output := CleanSlash(pairdir + "/" + name + "_to_" + ref + "ref.pairs")
	r.AddScripts(
		"mkdir -p `dirname $@`",
		CleanSlash(pt + " " + bam + " " + chrlens + " " + output),
		"touch $@",
	)

	mf.Add(r)

}

func AddRun(mf *makem.MakeData, p Params) {
	AddSplit(mf, p.Name, p.Nsplits, p.Indir, p.Outdir, p.Scriptdir)
	AddTrim(mf, p.Name, p.Nsplits, p.Outdir, p.Scriptdir)
	AddBwaRef(mf, p.Name, p.Ref, p.Refdir, p.Outdir, p.Scriptdir)
	AddBwa(mf, p.Name, p.Nsplits, p.Ref, p.Refdir, p.Outdir, p.Scriptdir)
	AddMerge(mf, p.Name, p.Nsplits, p.Outdir)
	AddFullBai(mf, p.Name, p.Outdir)
	AddPairtools(mf, p.Name, p.Refdir, p.Ref, p.Outdir, p.Scriptdir)
}

func MakeMakefile(params []Params) makem.MakeData {
	var mf makem.MakeData
	for _, p := range params {
		AddRun(&mf, p)
	}
	return mf
}

func MakeSplitPath(p Params, i int) string {
	return fmt.Sprintf("runscripts/run_%v_split_%04d.sh", p.Name, i)
}

func MakeSplit(p Params, i int) error {
	os.Mkdir("runscripts", 0777)
	path := MakeSplitPath(p, i)
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	target := CleanSlash(fmt.Sprintf("%v/%v_R1_001_split_%04d.bam.done", p.Outdir + "/bwa_split/", p.Name, i))

	fmt.Fprintf(
		w,
		`#!/bin/bash
set -e
#SBATCH -t 72:00:00    #max:    72 hours (24 on ash)
#SBATCH -N 1          #format: count or min-max
#SBATCH -A phadnis    #values: yandell, yandell-em (ember), ucgd-kp (kingspeak)
#SBATCH -p lonepeak    #kingspeak, ucgd-kp, kingspeak-freecycle, kingspeak-guest
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
	path := fmt.Sprintf("runscripts/run_%v_splits.sh", p.Name)
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	for i:=0; i<p.Nsplits; i++ {
		splitpath := MakeSplitPath(p, i)
		err = MakeSplit(p, i)
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "bash %v\n", splitpath)
	}
	return nil
}

func MakeEnds(p Params) error {
	path := fmt.Sprintf("runscripts/run_%v_end.sh", p.Name)
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	target := CleanSlash(p.Outdir + "/pairtools/pairtools_done.txt")

	fmt.Fprintf(
		w,
		`#!/bin/bash
set -e
#SBATCH -t 72:00:00    #max:    72 hours (24 on ash)
#SBATCH -N 1          #format: count or min-max
#SBATCH -A phadnis    #values: yandell, yandell-em (ember), ucgd-kp (kingspeak)
#SBATCH -p lonepeak    #kingspeak, ucgd-kp, kingspeak-freecycle, kingspeak-guest
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
	path := fmt.Sprintf("runscripts/run_%v_start.sh", p.Name)
	w, err := os.Create(path)
	if err != nil {
		return err
	}
	defer w.Close()

	target := CleanSlash(p.Outdir + "/bwa_split/reference.fa.done")

	fmt.Fprintf(
		w,
		`#!/bin/bash
set -e
#SBATCH -t 72:00:00    #max:    72 hours (24 on ash)
#SBATCH -N 1          #format: count or min-max
#SBATCH -A phadnis    #values: yandell, yandell-em (ember), ucgd-kp (kingspeak)
#SBATCH -p lonepeak    #kingspeak, ucgd-kp, kingspeak-freecycle, kingspeak-guest
#SBATCH -J %v_start        #Job name

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

func main() {
	params := []Params {
		Params {
			"axw",
			36,
			"axw",
			"/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic/data/axw_full/",
			"/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/refs/combos/axw/",
			"/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic2/out/axw_1",
			"scripts/",
		},
		Params {
			"ixl",
			38,
			"ixw",
			"/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic/data/ixl_full/",
			"/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/refs/combos/ixw/",
			"/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic2/out/ixl_1",
			"scripts/",
		},
		Params {
			"hxw",
			37,
			"ixw",
			"/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic/data/hxw_full/",
			"/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/refs/combos/ixw/",
			"/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic2/out/hxw_1",
			"scripts/",
		},
	}
	mf := MakeMakefile(params)
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
		err = MakeEnds(param)
		if err != nil {
			panic(err)
		}
	}
}
