package main

import (
	"fmt"
	"github.com/jgbaldwinbrown/fasttsv"
	"bufio"
	"io"
	"os"
	"flag"
)
type pairfile_stats struct {
	path string
	mm_count int
	uu_count int
	mu_count int
	um_count int
	ww_count int
	nn_count int
	xx_count int
	nm_count int
	nr_count int
	nu_count int
	mr_count int
	ru_count int
	ur_count int
	dd_count int
	good_count int
	bad_count int
}
type flag_holder struct {
	path_list string
}

func get_flags() flag_holder {
	var out flag_holder
	flag.StringVar(&out.path_list, "p", "", "List of paths to process (default stdin-only).")
	flag.Parse()
	return out
}

func print_stats(p pairfile_stats, w io.Writer) {
	fmt.Fprintf(w, "File: %s\n", p.path)
	fmt.Fprintf(w, "UU reads: %d\n", p.uu_count)
	fmt.Fprintf(w, "MM reads: %d\n", p.mm_count)
	fmt.Fprintf(w, "MU reads: %d\n", p.mu_count)
	fmt.Fprintf(w, "UM reads: %d\n", p.um_count)
	fmt.Fprintf(w, "WW reads: %d\n", p.ww_count)
	fmt.Fprintf(w, "NN reads: %d\n", p.nn_count)
	fmt.Fprintf(w, "XX reads: %d\n", p.xx_count)
	fmt.Fprintf(w, "NM reads: %d\n", p.nm_count)
	fmt.Fprintf(w, "NR reads: %d\n", p.nr_count)
	fmt.Fprintf(w, "NU reads: %d\n", p.nu_count)
	fmt.Fprintf(w, "MR reads: %d\n", p.mr_count)
	fmt.Fprintf(w, "MU reads: %d\n", p.mu_count)
	fmt.Fprintf(w, "RU reads: %d\n", p.ru_count)
	fmt.Fprintf(w, "UR reads: %d\n", p.ur_count)
	fmt.Fprintf(w, "DD reads: %d\n", p.dd_count)
	fmt.Fprintf(w, "good reads: %d\n", p.good_count)
	fmt.Fprintf(w, "bad reads: %d\n", p.bad_count)
}

func print_all_stats(ps []pairfile_stats, w io.Writer) {
	for _, p := range ps {
		fmt.Fprintf(w, "---\n")
		print_stats(p, w)
		fmt.Fprintf(w, "---\n")
	}
}

func pair_stats(path string, r io.Reader) pairfile_stats {
	var out pairfile_stats
	out.path = path
	bscanner := bufio.NewScanner(r)
	for bscanner.Scan() {
		if len(bscanner.Text()) >= 9 && bscanner.Text()[:9] == "#columns:" { break }
	}
	scanner := fasttsv.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Line())
		if len(scanner.Line()) >= 8 {
			code := scanner.Line()[7]
			switch code {
			case "MM":
				out.mm_count++
				out.bad_count++
			case "UU":
				out.uu_count++
				out.good_count++
			case "MU":
				out.mu_count++
				out.bad_count++
			case "UM":
				out.um_count++
				out.bad_count++
			case "WW":
				out.ww_count++
				out.bad_count++
			case "NN":
				out.nn_count++
				out.bad_count++
			case "XX":
				out.xx_count++
				out.bad_count++
			case "NM":
				out.nm_count++
				out.bad_count++
			case "NR":
				out.nr_count++
				out.bad_count++
			case "NU":
				out.nu_count++
				out.bad_count++
			case "MR":
				out.mr_count++
				out.bad_count++
			case "RU":
				out.ru_count++
				out.good_count++
			case "UR":
				out.ur_count++
				out.good_count++
			case "DD":
				out.dd_count++
				out.bad_count++
			default:
			}
		}
	}
	return out
}

func all_pair_stats(path_list_path string) []pairfile_stats {
	var out []pairfile_stats
	path_list_conn, err := os.Open(path_list_path)
	if err != nil {panic(err)}
	s := bufio.NewScanner(path_list_conn)
	for s.Scan() {
		conn, err := os.Open(s.Text())
		if err != nil {panic(err)}
		stats := pair_stats(s.Text(), conn)
		out = append(out, stats)
		conn.Close()
	}
	path_list_conn.Close()
	return out
}

func main() {
	flags := get_flags()
	if flags.path_list == "" {
		stats := pair_stats("", os.Stdin)
		print_stats(stats, os.Stdout)
	} else {
		all_stats := all_pair_stats(flags.path_list)
		print_all_stats(all_stats, os.Stdout)
	}
}
