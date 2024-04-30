package prepfa

import (
	"bufio"
	"os/signal"
	"syscall"
	"github.com/jgbaldwinbrown/csvh"
	"context"
	"fmt"
	"os/exec"
	"encoding/json"
	"golang.org/x/sync/errgroup"
	"os"
	"io"
	"log"
	"flag"
)

type TensorflowFlags struct {
	Fa1 string
	Fa2 string
	Bed string
	Paircol int
	Width int
}

type CrossArgs struct {
	Fa string
	Vcf string
	Bed string
	C0 int
	C1 int
	Parent string
}

type Args struct {
	Crosses []CrossArgs
	Size int
	Step int
	Width int
	Outpre string
	Threads int
	Paircol int
	Prepare bool
	Cleanup bool
	Combine bool
	Predict bool
}

func (a Args) CrossPrepFlags(i int) PrepFaFlags {
	return PrepFaFlags{
		Fa: a.Crosses[i].Fa,
		Vcf: a.Crosses[i].Vcf,
		C0: a.Crosses[i].C0,
		C1: a.Crosses[i].C1,
		Size: a.Size,
		Step: a.Step,
		Width: a.Width,
		Outpre: fmt.Sprintf("%v_prep_%v", a.Outpre, i),
	}
}

func (a Args) CrossCleanupFlags(i int) CleanupFlags {
	return CleanupFlags {
		Fa1: fmt.Sprintf("%v_prep_%v_wins_1.fa.gz", a.Outpre, i),
		Fa2: fmt.Sprintf("%v_prep_%v_wins_2.fa.gz", a.Outpre, i),
		Bed: a.Crosses[i].Bed,
		Outpre: fmt.Sprintf("%v_cleanup_%v", a.Outpre, i),
		Parent: a.Crosses[i].Parent,
		Paircol: a.Paircol,
	}
}

func (a Args) TensorflowArgs() TensorflowFlags {
	fa1path := a.Outpre + "_1.fa"
	fa2path := a.Outpre + "_2.fa"
	bedpath := a.Outpre + ".bed"
	return TensorflowFlags {
		Fa1: fa1path,
		Fa2: fa2path,
		Bed: bedpath,
		Paircol: a.Paircol,
		Width: a.Width,
	}
}

func StartSignalhandler(closef func()) (endf func()) {
	sigc := make(chan os.Signal, 4)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	done := make(chan struct{})
	endf = func() {
		close(done)
	}

	go func() {
		select {
		case s := <-sigc:
			switch s {
			case syscall.SIGHUP:
				fallthrough
			case syscall.SIGINT:
				fallthrough
			case syscall.SIGTERM:
				fallthrough
			case syscall.SIGQUIT:
				closef()
				panic(fmt.Errorf("Canceled"))
			default:
			}
		case <-done:
		}
	}()
	return endf
}

func PrepareCtx(ctx context.Context, arg Args, i int) error {
	return PrepFa(arg.CrossPrepFlags(i))
}

func CleanupCtx(ctx context.Context, arg Args, i int) error {
	return Cleanup(arg.CrossCleanupFlags(i))
}

func RunSeparate(ctx context.Context, arg Args, i int) error {
	if arg.Prepare {
		if e := PrepareCtx(ctx, arg, i); e != nil {
			return e;
		}
	}
	if arg.Cleanup {
		if e := CleanupCtx(ctx, arg, i); e != nil {
			return e
		}
	}
	return nil
}

func RunSeparates(ctx context.Context, args Args) error {
	g, ctx2 := errgroup.WithContext(ctx)
	if args.Threads > 0 {
		g.SetLimit(args.Threads)
	}

	for i, _ := range args.Crosses {
		i := i
		g.Go(func() error {
			return RunSeparate(ctx2, args, i)
		})
	}

	return g.Wait()
}

func CatPath(w io.Writer, path string) error {
	r, e := csvh.OpenMaybeGz(path)
	if e != nil {
		return e
	}
	defer r.Close()
	br := bufio.NewReader(r)
	_, e = io.Copy(w, br)
	return e
}

func CatPaths(out string, paths ...string) error {
	w, e := csvh.CreateMaybeGz(out)
	if e != nil {
		return e
	}
	defer w.Close()
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	for _, path := range paths {
		if e := CatPath(bw, path); e != nil {
			return e
		}
	}
	return nil
}

func TensorflowPredict(ctx context.Context, w io.Writer, f TensorflowFlags) error {
	cmd := exec.CommandContext(
		ctx,
		"tensorflow_predict.py",
		f.Fa1,
		f.Fa2,
		f.Bed,
		fmt.Sprint(f.Paircol),
		fmt.Sprint(f.Width),
	)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// tensorflow_predict.py \
// 	inputs2_1.fa \
// 	inputs2_2.fa \
// 	inputs2.bed \
// 	7 \
// 	90000 | \
// pigz -p 8 > predicted.txt.gz

func Join(ctx context.Context, args Args) error {
	if args.Combine {
		fa1s := make([]string, 0, len(args.Crosses))
		fa2s := make([]string, 0, len(args.Crosses))
		beds := make([]string, 0, len(args.Crosses))
		for i, _ := range args.Crosses {
			cf := args.CrossCleanupFlags(i)
			fa1s = append(fa1s, cf.Outpre + "_1.fa.gz")
			fa2s = append(fa2s, cf.Outpre + "_2.fa.gz")
			beds = append(beds, cf.Outpre + ".bed.gz")
		}

		fa1path := args.Outpre + "_1.fa"
		fa2path := args.Outpre + "_2.fa"
		bedpath := args.Outpre + ".bed"
		if e := CatPaths(fa1path, fa1s...); e != nil {
			return e
		}
		if e := CatPaths(fa2path, fa2s...); e != nil {
			return e
		}
		if e := CatPaths(bedpath, beds...); e != nil {
			return e
		}
	}

	if args.Predict {
		targ := args.TensorflowArgs()
		outp := args.Outpre + "_out.txt.gz"
		w, e := csvh.CreateMaybeGz(outp)
		if e != nil {
			return e
		}
		defer w.Close()
		bw := bufio.NewWriter(w)
		defer bw.Flush()
		return TensorflowPredict(ctx, bw, targ)
	}
	return nil
}

func GetArgs(r io.Reader) (Args, error) {
	var a Args
	dec := json.NewDecoder(r)
	e := dec.Decode(&a)
	return a, e
}

type FullFlags struct {
	NoPrepare bool
	NoCleanup bool
	NoPredict bool
	NoCombine bool
	PredictOnly bool
}

func RunPrepAndClean() {
	var f FullFlags
	flag.BoolVar(&f.PredictOnly, "p", false, "Only run prediction, assume setup is already complete")
	flag.BoolVar(&f.NoPrepare, "noprep", false, "Do not run prepare step")
	flag.BoolVar(&f.NoCleanup, "noclean", false, "Do not run cleanup step")
	flag.BoolVar(&f.NoCombine, "nocombine", false, "Do not run combining step")
	flag.BoolVar(&f.NoPredict, "nopredict", false, "Do not run prediction step")
	flag.Parse()

	args, e := GetArgs(os.Stdin)
	if e != nil {
		log.Fatal(e)
	}
	args.Prepare = !f.NoPrepare
	args.Cleanup = !f.NoCleanup
	args.Predict = !f.NoPredict

	if f.PredictOnly {
		args.Prepare = false
		args.Cleanup = false
		args.Combine = false
	}

	ctx, closef := context.WithCancel(context.Background())
	sigend := StartSignalhandler(closef)
	defer sigend()

	if args.Prepare || args.Cleanup {
		if e = RunSeparates(ctx, args); e != nil {
			log.Fatal(e)
		}
	}

	if args.Predict || args.Combine {
		if e = Join(ctx, args); e != nil {
			log.Fatal(e)
		}
	}
}
