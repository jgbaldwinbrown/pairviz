package register

import (
	"context"
	"os/exec"
	"fmt"
	"io"
	"encoding/json"
	"flag"
	"os"
	"golang.org/x/sync/errgroup"
	"bufio"
)

type Job struct {
	Inpath string
	Outpath string
	Plot bool
	Plotoutpath string
	Plotname string
	Mindist int64
	Maxdist int64
}

func Must(e error) {
	if e != nil {
		panic(e)
	}
}

func RunPaths(inpath, outpath string) error {
	r, e := OpenMaybeGz(inpath)
	if e != nil {
		return e
	}
	defer func() { Must(r.Close()) }()
	br := bufio.NewReader(r)

	w, e := CreateMaybeGz(outpath)
	if e != nil {
		return e
	}
	defer func() { Must(w.Close()) }()
	bw := bufio.NewWriter(w)
	defer func() { Must(bw.Flush()) }()

	return Run(br, bw)
}

func RunPlot(ctx context.Context, j Job) error {
	cmd := exec.CommandContext(
		ctx,
		"plotregisters",
		j.Outpath, j.Plotoutpath, j.Plotname,
		fmt.Sprintf("%v", j.Mindist),
		fmt.Sprintf("%v", j.Maxdist),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunJob(ctx context.Context, j Job) error {
	e := RunPaths(j.Inpath, j.Outpath)
	if e != nil {
		return e
	}

	if !j.Plot {
		return nil
	}

	return RunPlot(ctx, j)
}

func RegisterMulti(ctx context.Context, threads int, jobs ...Job) error {
	g, ctx2 := errgroup.WithContext(ctx)
	if threads > 0 {
		g.SetLimit(threads)
	}
	for _, job := range jobs {
		job := job
		g.Go(func() error {
			return RunJob(ctx2, job)
		})
	}
	err := g.Wait()
	return err
}

func FullRegisterMulti() {
	threads := flag.Int("t", -1, "Threads to use (default infinite).")
	flag.Parse()

	dec := json.NewDecoder(os.Stdin)
	var jobs []Job
	var j Job
	for e := dec.Decode(&j); e != io.EOF; e = dec.Decode(&j) {
		Must(e)
		jobs = append(jobs, j)
	}

	e := RegisterMulti(context.Background(), *threads, jobs...)
	Must(e)
}
