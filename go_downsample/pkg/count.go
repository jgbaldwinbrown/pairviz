package dsamp

import (
	"fmt"
	"io"
	"bufio"
	"github.com/jgbaldwinbrown/lscan/pkg"
	"regexp"
	"compress/gzip"
	"os"
)

var comment = regexp.MustCompile(`^#`)

func IsComment(line string) bool {
	return comment.MatchString(line)
}

var uniqueSplit = lscan.ByByte('\t')

func IsUnique(str string, linebuf []string) bool {
	linebuf = lscan.SplitByFunc(linebuf, str, uniqueSplit)
	return len(linebuf) == 8 && linebuf[1] != "!" && linebuf[3] != "!"
}

func CountPairtoolsUnique(r io.Reader) (int, error) {
	s := bufio.NewScanner(r)
	s.Buffer([]byte{}, 1e12)
	count := 0
	var line []string
	for s.Scan() {
		if !IsComment(s.Text()) {
			if IsUnique(s.Text(), line) {
				count++
			}
		}
	}
	return count, nil
}

func CountPairtoolsUniqueGzfile(path string) (int, error) {
	r, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	gr, err := gzip.NewReader(r)
	if err != nil {
		return 0, err
	}
	defer gr.Close()

	return CountPairtoolsUnique(gr)
}

func CountPairtoolsUniqueGzfiles(paths ...string) ([]int, error) {
	var counts []int
	for _, path := range paths {
		count, err := CountPairtoolsUniqueGzfile(path)
		if err != nil {
			return counts, err
		}
		counts = append(counts, count)
	}
	return counts, nil
}

func GetPaths(r io.Reader) []string {
	var paths []string
	s := bufio.NewScanner(r)
	for s.Scan() {
		paths = append(paths, s.Text())
	}
	return paths
}

func PrintPathCounts(w io.Writer, paths []string, counts []int) {
	for i, path := range paths {
		if len(counts) > i {
			fmt.Fprintf(w, "%v\t%v\n", counts[i], path)
		}
	}
}

func RunCounts() {
	paths := GetPaths(os.Stdin)
	counts, err := CountPairtoolsUniqueGzfiles(paths...)
	if err != nil {
		panic(err)
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	PrintPathCounts(w, paths, counts)
}
