package main

import (
	"os"
	"testing"
	"strings"
)

var gTestIn = `## pairs format v1.0.0
#shape: upper triangle
#genome_assembly: unknown
#chromsize: X_ISO1 23542271
#chromsize: X_W501 22038717
#chromsize: 2L_ISO1 23513712
#chromsize: 2L_W501 25159701
#chromsize: 2R_ISO1 25286936
#chromsize: 2R_W501 22319944
#chromsize: 3L_ISO1 28110227
#chromsize: 3L_W501 23401300
#chromsize: 3R_ISO1 32079331
#chromsize: 3R_W501 28152764
#chromsize: 4_ISO1 1348131
#chromsize: 4_W501 1146858
#samheader: @HD	VN:1.6	SO:queryname
#samheader: @SQ	SN:X_ISO1	LN:23542271
#samheader: @SQ	SN:X_W501	LN:22038717
#samheader: @SQ	SN:2L_ISO1	LN:23513712
#samheader: @SQ	SN:2L_W501	LN:25159701
#samheader: @SQ	SN:3L_ISO1	LN:28110227
#samheader: @SQ	SN:3L_W501	LN:23401300
#samheader: @SQ	SN:3R_ISO1	LN:32079331
#samheader: @SQ	SN:3R_W501	LN:28152764
#samheader: @SQ	SN:2R_ISO1	LN:25286936
#samheader: @SQ	SN:2R_W501	LN:22319944
#samheader: @SQ	SN:4_ISO1	LN:1348131
#samheader: @SQ	SN:4_W501	LN:1146858
#samheader: @PG	ID:bwa	PN:bwa	VN:0.7.15-r1140	CL:bwa mem -t 32 /data1/jbrown/drosophila_homologous_joining_project/temp/hic/ixw_1/bwa//reference.fa /dev/fd/63 /dev/fd/62
#samheader: @PG	ID:pairtools_parse	PN:pairtools_parse	CL:/data1/jbrown/local_programs/anaconda/install_dir/anaconda/bin/pairtools parse -c /data1/jbrown/drosophila_homologous_joining_project/raw_data/refs/combos/ixw/ixw.fa.gz.chrlens.txt -o temp/hic/ixw_1/pairtools/ixw_to_ixwref.pairs --drop-sam /data1/jbrown/drosophila_homologous_joining_project/temp/hic/ixw_1/bwa/15818X2_190307_D00550_0549_BCD5KUANXX_S402_L005_filtered.bam	PP:bwa	VN:0.2.2
#columns: readID chrom1 pos1 chrom2 pos2 strand1 strand2 pair_type
D00550:549:CD5KUANXX:5:1101:1126:2209	X_ISO1	5	X_ISO1	6	+	-	UU
D00550:549:CD5KUANXX:5:1101:1126:2209	X_W501	5	X_ISO1	6	+	-	UU
D00550:549:CD5KUANXX:5:1101:1126:2209	2L_W501	9	2L_ISO1	9	+	-	UU
D00550:549:CD5KUANXX:5:1101:1126:2209	X_W501	5	X_ISO1	6	-	+	UU
D00550:549:CD5KUANXX:5:1101:1126:2209	2L_W501	9	2L_ISO1	9	+	-	UU`


var gFlags = Flags {
	WinSize: 10,
	WinStep: 3,
	Distance: 100000,
	Chromosome: false,
	Name: "test5",
	NameCol: true,
	Stdin: true,
	NoFpkm: false,
	Region: "",
	SeparateGenomes: true,
	MinDistance: -1,
	PairMinDistance: -1,
	SelfInMinDistance: -1,
	ReadLen: 150,
}

func TestFull(t *testing.T) {
	flags := gFlags
	in := strings.NewReader(gTestIn)
	// var out strings.Builder

	FprintWinStats(os.Stdout, WinStats(flags, in), flags.SeparateGenomes, flags.ReadLen)
}

func TestFullSelfInMin(t *testing.T) {
	flags := gFlags
	flags.SelfInMinDistance = 3
	in := strings.NewReader(gTestIn)
	FprintWinStats(os.Stdout, WinStats(flags, in), flags.SeparateGenomes, flags.ReadLen)
}
