package colorinter

import (
	"fmt"
	"flag"
	"os"
	"os/exec"
	"io"
)

type Cols struct {
	Chrom int
	Start int
	End int
	Val int
}

type ColsPair struct {
	A Cols
	B Cols
}

// Intersect a bed file of hits to a gff file of features using bedtools' intersect -wao settings
func Intersect(hitsbed, colorsgff string, w io.Writer) (*exec.Cmd, error) {
	cmd := exec.Command("bedtools", "intersect", "-wao", "-a", hitsbed, "-b", colorsgff)
	cmd.Stderr = os.Stderr
	cmd.Stdout = w
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

// Plot the output of Intersect
func PlotWhisker(intersectbed string, valcol, colorcol int, outpath string) error {
	cmd := exec.Command("plot_color_whisker", intersectbed, fmt.Sprintf("%d", valcol), fmt.Sprintf("%d", colorcol), outpath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func RunIntersect() {
	bedpath := flag.String("i", "", "Input path (required)")
	gff := flag.String("g", "", "Gff or bed containing color identities (required)")
	outpre := flag.String("o", "", "Output prefix (required)")
	valcol := flag.Int("v", -1, "Value column (required)")
	colorcol := flag.Int("c", -1, "Color column (required)")
	flag.Parse()

	if *bedpath == "" {
		panic(fmt.Errorf("-i missing"))
	}
	if *gff == "" {
		panic(fmt.Errorf("-g missing"))
	}
	if *outpre == "" {
		panic(fmt.Errorf("-o missing"))
	}
	if *valcol == -1 {
		panic(fmt.Errorf("-v missing"))
	}
	if *colorcol == -1 {
		panic(fmt.Errorf("-c missing"))
	}

	bedoutpath := *outpre + ".bed"
	pngoutpath := *outpre + ".png"

	bedout, err := os.Create(bedoutpath)
	if err != nil {
		panic(err)
	}
	defer bedout.Close()

	cmd, err := Intersect(*bedpath, *gff, bedout)
	if err != nil {
		panic(err)
	}
	err = cmd.Wait()
	if err != nil {
		panic(err)
	}

	err = PlotWhisker(bedoutpath, *valcol, *colorcol, pngoutpath)
	if err != nil {
		panic(err)
	}
}

/*
func CountColorsFulls(r io.Reader, cols ColsPair) map[string]float64 {
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)
	var line []string
	split := lscan.ByByte('\t')
	out := map[string][]float64{".":0}
	for s.Scan() {
		line = lscan.SplitByFunc(line, s.Text(), split)
		color := line[cols.B.Val]
		val := GetVal(line[cols.A.Val])
		if _, ok := out[color]; !ok {
			out[color] = 0;
		}
		out[color]
	}
}
*/

/*
Jim@T480:go_intersect$ bedtools intersect -wao -a vals.bed -b colors.bed 
chr1	55	336	0.01	chr1	0	329	green	274
chr1	55	336	0.01	chr1	329	333	black	4
chr1	55	336	0.01	chr1	333	399	yellow	3
chr1	330	399	0.5	chr1	329	333	black	3
chr1	330	399	0.5	chr1	333	399	yellow	66
chr2	1	1000	0.9	chr2	1	1000	blue	999
chr1	1000	10000	333.33	.	-1	-1	.	0
*/
