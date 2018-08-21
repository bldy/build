// Copyright 2018 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package main // import "bldy.build/build/cmd/bldy"

import (
	"os"

	_ "bldy.build/build/cmd/build"
	"bldy.build/build/cmd/internal/cmds"
	_ "bldy.build/build/cmd/query"
	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Name:     "bldy",
		Usage:    "build things concurrently",
		Commands: cmds.Commands(),
	}
	app.Run(os.Args)
}