package build

import (
	"context"
	"math"
	"os"
	"runtime"

	"bldy.build/build/builder"
	"bldy.build/build/cmd/internal/cmds"
	"bldy.build/build/graph"
	cli "gopkg.in/urfave/cli.v2"
)

func init() {
	cmds.RegisterCommand(&cli.Command{
		Name: "build",
		Flags: []cli.Flag{
			&cli.BoolFlag{Name: "fresh", Value: false, Usage: "use the cache"},
		},
		Aliases: []string{"b"},
		Usage:   "builds a target",
		Action:  buildMain,
	})
}

func buildMain(c *cli.Context) error {
	args := c.Args()
	if args.Len() < 1 {
		return cli.Exit("build requires atleast 1 argument", 1)
	}
	wd, err := os.Getwd()
	if err != nil {
		return cli.Exit(err.Error(), 3)
	}
	g, err := graph.New(wd, args.Get(0))
	if err != nil {
		return cli.Exit(err.Error(), 4)
	}
	if g == nil {
		return cli.Exit("we could not construct your graph", 1)
	}
	workers := float64(runtime.NumCPU()) * 1.25
	b := builder.New(
		g,
		&builder.Config{
			Fresh: c.Bool("fresh"),
		},
		newNotifier(int(math.Round(workers))),
	)
	ctx := context.Background()
	b.Execute(ctx, int(math.Round(workers)))

	if err != nil {
		return cli.Exit(err.Error(), 2)
	}

	return nil
}
