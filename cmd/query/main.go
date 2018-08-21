package build

import (
	"fmt"
	"os"

	"bldy.build/build/cmd/internal/cmds"
	"bldy.build/build/graph"
	cli "gopkg.in/urfave/cli.v2"
	"sevki.org/x/pretty"
)

func init() {
	cmds.RegisterCommand(&cli.Command{
		Name:    "query",
		Aliases: []string{"q"},
		Usage:   "prints a target",
		Action:  queryMain,
	})
	cmds.RegisterCommand(&cli.Command{
		Name:    "deps",
		Aliases: []string{"d"},
		Usage:   "prints dependencies of a target",
		Action:  depsMain,
	})
}

func queryMain(c *cli.Context) error {
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
	fmt.Println(pretty.JSON(g.Root.Target))
	return nil
}
func depsMain(c *cli.Context) error {
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
	fmt.Println(pretty.JSON(g.Root.Target.Dependencies()))
	return nil
}
