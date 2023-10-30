package main

import (
	"bufio"
	"strings"
	"encoding/csv"
	"io"
	"os"
	"fmt"
	"path/filepath"
)

type Input struct {
	OldR1 string
	OldR2 string
	Basedir string
	Comboname string
}

func LinkOne(i Input) error {
	outdir := filepath.Join(i.Basedir, fmt.Sprintf("%v_full", i.Comboname))

	if err := os.MkdirAll(outdir, 0777); err != nil {
		return err
	}

	outr1 := filepath.Join(outdir, fmt.Sprintf("%v_R1_001.fastq.gz", i.Comboname))
	if err := os.Symlink(i.OldR1, outr1); err != nil {
		return err
	}

	outr2 := filepath.Join(outdir, fmt.Sprintf("%v_R2_001.fastq.gz", i.Comboname))
	if err := os.Symlink(i.OldR2, outr2); err != nil {
		return err
	}

	return nil
}

func LinkAll(is []Input) error {
	for _, i := range is {
		if err := LinkOne(i); err != nil {
			return err
		}
	}
	return nil
}

func Scan(line []string, ptrs ...any) error {
	for i, ptr := range ptrs {
		if s, ok := ptr.(*string); ok {
			*s = line[i]
		} else {
			_, e := fmt.Sscanf(line[i], "%v", ptr)
			if e != nil {
				return e
			}
		}
	}
	return nil
}

func FindPath(id, suffix string, paths []string) (string, error) {
	idu := id + "_"
	for _, path := range paths {
		if strings.Contains(path, idu) && strings.Contains(path, suffix) {
			return path, nil
		}
	}
	return "", fmt.Errorf("FindPath could not find %v, %v", idu, suffix)
}

var NameMapOld = map[string]string {
	"ISO1 X A4": "ixa4",
	"ISO1 X A7": "ixa7",
	"Nueva X w501": "nxw",
	"M252 X w501": "mxw",
	"ISO1 X w501": "ixw",
	"A7 X Nueva": "a7xn",
	"Hmr X w501": "hxw",
	"ISO1 X Lhr": "ixl",
	"ISO1 X Sawamura": "ixs",
	"Sawamura X w501": "sxw",
	"Salivary gland": "sal",
	"Head + thorax": "adult",
	"Brain + imaginal discs": "brain",
	"Fat body": "fat",
}

var NameMap = map[string]string {
	"ISO1 X A4": "iso1xa4",
	"ISO1 X A7": "iso1xa7",
	"Nueva X w501": "s14xw501",
	"M252 X w501": "m252xw501",
	"ISO1 X w501": "iso1xw501",
	"A7 X Nueva": "a7xs14",
	"Hmr X w501": "hmrxw501",
	"ISO1 X Lhr": "iso1xlhr",
	"ISO1 X Sawamura": "iso1xsaw",
	"Sawamura X w501": "sawxw501",
	"Salivary gland": "sal",
	"Head + thorax": "adult",
	"Brain + imaginal discs": "brain",
	"Fat body": "fat",
}

func ParseInput(line []string, paths []string) (Input, error) {
	geno := NameMap[line[4]]
	tissue := NameMap[line[5]]

	var i Input
	var e error
	if i.OldR1, e = FindPath(line[0], "_R1_001.fastq.gz", paths); e != nil {
		return i, e
	}
	if i.OldR2, e = FindPath(line[0], "_R2_001.fastq.gz", paths); e != nil {
		return i, e
	}

	i.Basedir = "/uufs/chpc.utah.edu/common/home/shapiro-group3/jim/new/fly/hic4_final/data/21326R"
	i.Comboname = fmt.Sprintf("%v_%v", geno, tissue)

	return i, nil
}

func ReadLines(path string) ([]string, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	s := bufio.NewScanner(r)
	var out []string
	for s.Scan() {
		if s.Err() != nil {
			return nil, s.Err()
		}
		out = append(out, s.Text())
	}
	return out, nil
}

func MakeInputs(r io.Reader) ([]Input, error) {
	paths, err := ReadLines("paths.txt")
	if err != nil {
		return nil, err
	}

	cr := csv.NewReader(r)
	cr.Comma = rune('\t')
	cr.FieldsPerRecord = -1
	cr.ReuseRecord = true
	cr.LazyQuotes = true

	var out []Input

	for l, e := cr.Read(); e != io.EOF; l, e = cr.Read() {
		if e != nil {
			return nil, e
		}
		i, e := ParseInput(l, paths)
		if e != nil {
			fmt.Println(e)
			continue
		}
		out = append(out, i)
	}
	return out, nil
}

func main() {
	is, err := MakeInputs(os.Stdin)
	if err != nil {
		panic(err)
	}

	if err = LinkAll(is); err != nil {
		panic(err)
	}
}
