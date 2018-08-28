package build

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"

	"bldy.build/build/builder"
	"bldy.build/build/graph"
	"bldy.build/build/label"
	"github.com/google/subcommands"
)

type BuildCmd struct {
	fresh bool
}

func (*BuildCmd) Name() string     { return "build" }
func (*BuildCmd) Synopsis() string { return "builds a target" }
func (*BuildCmd) Usage() string {
	return `build //<package>:<name>
Builds a target
`
}

func (b *BuildCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&b.fresh, "fresh", false, "use the cache or build fresh")
}

func (b *BuildCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if len(args) != 1 {
		return subcommands.ExitUsageError
	}
	l, ok := args[0].(label.Label)
	if !ok {
		return subcommands.ExitUsageError
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		return 3
	}
	g, err := graph.New(wd, string(l))
	if err != nil {
		fmt.Println(err.Error())
		return 4
	}
	if g == nil {
		fmt.Println("nothing to build")
		return 5
	}
	workers := float64(runtime.NumCPU()) * 1.25
	bldr := builder.New(
		g,
		&builder.Config{
			Fresh: b.fresh,
		},
		newNotifier(int(math.Round(workers))),
	)
	bldr.Execute(ctx, int(math.Round(workers)))

	if err != nil {
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
