package main

import (
    "fmt"
    "flag"
    "bufio"
    "io"
    "os"
    "log"
    "strconv"
    "strings"
)

type flaglist struct {
    width int64
    step int64
    centers_path string
    centers []pos_entry
}

type pos_entry struct {
    chr string
    pos int64
}

type radius_table struct {
    entries []radius_entry
    width int64
    step int64
}

type radius_entry struct {
    pair_count int64
    self_count int64
    uninform_count int64
}

type pair_entry struct {
    name string
    species1 string
    chr1 string
    pos1 int64
    dir1 byte
    species2 string
    chr2 string
    pos2 int64
    dir2 byte
    unambiguous bool
    hybrid bool
}

func get_centers(path string) []pos_entry {
    centers := make([]pos_entry, 0)
    var p pos_entry
    file, err := os.Open(path)
    if err != nil {
        log.Fatal(err)
    }
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        sl := strings.Split(strings.TrimSuffix(scanner.Text(), "\n"), "\t")
        p.chr = sl[0]
        pos, ierr := strconv.Atoi(sl[1])
        p.pos = int64(pos)
        if ierr != nil {
            log.Fatal(ierr)
        }
        centers = append(centers, p)
    }
    return centers;
}

func get_flags() flaglist {
    var flags flaglist;
    var width_temp int;
    var step_temp int;
    flag.IntVar(&width_temp, "width", 1000, "The width of intervals for calculating pairing at range.")
    flag.IntVar(&step_temp, "step", 100, "The step distance of intervals for calculating pairing at range.")
    flag.StringVar(&flags.centers_path, "center", "", "The central point around which to calculate pairing rates.")
    flag.Parse()
    if flags.centers_path == "" {
        fmt.Fprintln(os.Stderr, "A path to a set of centers is required.")
    }
    flags.centers = get_centers(flags.centers_path)
    flags.width = int64(width_temp);
    flags.step = int64(step_temp);
    return flags
}

func dist(a int64, b int64) int64 {
    if a<b {
        return b-a
    }
    return a-b
}

func closer(pos pos_entry, center1 pos_entry, center2 pos_entry) bool {
    if center1.chr != pos.chr {
        return false
    } else {
        if center2.chr != pos.chr {
            return true
        } else {
            return dist(pos.pos, center1.pos) < dist(pos.pos, center2.pos)
        }
    }
    return false
}

func find_closest(pos pos_entry, centers []pos_entry) pos_entry {
    best_center := pos_entry{chr:"", pos:-100e9}
    for _, center := range centers {
        if best_center.pos < 0 || closer(pos, center, best_center) {
            best_center = center
        }
    }
    return best_center
}

func make_radius_table(width int64, step int64) radius_table {
    return radius_table{
        width: width,
        step: step,
        entries: make([]radius_entry, 0),
    }
}

func parse_pairfile_line(line string) pair_entry {
    var p pair_entry
    var err error
    var pos1 int
    var pos2 int
    sl := strings.Fields(strings.TrimSuffix(line, "\n"))
    sc1_split := strings.Split(sl[1], "_")
    sc2_split := strings.Split(sl[3], "_")
    p.name = sl[0]
    p.species1 = sc1_split[0]
    p.chr1 = sc1_split[1]
    pos1, err = strconv.Atoi(sl[2])
    p.pos1 = int64(pos1)
    if err != nil {
        p.pos1 = -1
    }
    p.species2 = sc2_split[0]
    p.chr2 = sc2_split[1]
    pos2, err = strconv.Atoi(sl[4])
    p.pos2 = int64(pos2)
    if err != nil {
        p.pos2 = -1
    }
    p.dir1 = sl[5][0]
    p.dir2 = sl[6][0]
    p.hybrid = p.species1 != p.species2
    p.unambiguous = sl[7] == "UU"
    return p
}

func write_radius_table(table radius_table, outconn io.Writer) {
    
}

func pairing_radius(flags flaglist, inconn io.Reader, outconn io.Writer) {
    scanner := bufio.NewScanner(inconn)
    scanner.Buffer(make([]byte, 0), 8e9)
    output := make_radius_table(flags.width, flags.step)
    for scanner.Scan() {
        if len(scanner.Text()) > 0 && scanner.Text()[0] != '#' {
            pair_data := parse_pairfile_line(scanner.Text())
            add_pair_to_radius_table(pair_data, output)
        }
    }
    write_radius_table(output, outconn)
}

func main() {
    flags := get_flags()
    pairing_radius(flags, os.Stdin, os.Stdout)
}
