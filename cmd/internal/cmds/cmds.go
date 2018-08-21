// Copyright 2018 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmds

import cli "gopkg.in/urfave/cli.v2"

var commands []*cli.Command

func RegisterCommand(c *cli.Command) {
	commands = append(commands, c)
}
func Commands() []*cli.Command {
	return commands
}