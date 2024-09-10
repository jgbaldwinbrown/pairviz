package dsamp

import (
	"regexp"
	"fmt"
	"bufio"
	"os/exec"
	"os"
	"strings"
	"strconv"
)

// Count the number of uniquely-mapping reads in a bam file
func CountUniqueBam(path string) (count int64, err error) {
	var buf strings.Builder
	cmd := exec.Command("samtools", "view", "-c", "-F0x04", path)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		return 0, fmt.Errorf("CountUniqueBam: error running %w", err)
	}
	output := buf.String()

	cleanre := regexp.MustCompile(`[\n \t]`)
	count, err = strconv.ParseInt(cleanre.ReplaceAllString(output, ""), 0, 64)
	if err != nil {
		return 0, fmt.Errorf("CountUniqueBam: error parsing output %v %w", output, err)
	}
	return count, nil
}

// Count the uniquely-mapping reads in a set of bam files
func CountUniqueBams(paths ...string) (counts []int, err error) {
	for _, path := range paths {
		var count int64
		count, err = CountUniqueBam(path)
		if err != nil {
			return counts, fmt.Errorf("Count Unique Bams: counts: %v; err: %w", counts, err)
		}
		counts = append(counts, int(count))
	}
	return counts, nil
}

// Subset a bam file using the provided inpath, outpuath, seed, and proportion, only keeping unique reads
func SubsetBamUniques(arg SubsetArgs) error {
	w, err := os.Create(arg.Outpath)
	if err != nil {
		return err
	}
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	re := regexp.MustCompile(`.*\.`)
	subsetcmdarg := fmt.Sprintf("%d.%s", arg.Seed, re.ReplaceAllString(fmt.Sprintf("%f", float64(arg.Prop)), ""))
	cmd := exec.Command("samtools", "view", "-bS", "-F0x04", "-s", subsetcmdarg, arg.Inpath)
	cmd.Stdout = bw
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Subset multiple bam files, only keeping unique reads
func SubsetBamsUniques(args ...SubsetArgs) error {
	for _, arg := range args {
		err := SubsetBamUniques(arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func RunSubsetBamsUniques() {
	args := GetArgs()
	var inpaths []string
	for _, set := range args.IoSets {
		inpaths = append(inpaths, set.Inpath)
	}
	counts, err := CountUniqueBams(inpaths...)
	if err != nil {
		panic(err)
	}

	countw, err := os.Create(args.Countspath)
	if err != nil {
		panic(err)
	}
	defer countw.Close()
	PrintPathCounts(countw, inpaths, counts)

	props := GetLowestProps(counts)

	subargs := MakeSubsetArgs(args, props)
	err = SubsetBamsUniques(subargs...)
	if err != nil {
		panic(err)
	}
}
