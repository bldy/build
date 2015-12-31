// Copyright 2015 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// package term handles terminal display statuses
package term // import "sevki.org/build/term"

import (
	"fmt"
	"time"

	tm "github.com/buger/goterm"
	"github.com/fatih/color"
	"sevki.org/build/builder"
	nstime "sevki.org/lib/time"
)

var (
	statuses    map[int]builder.Update
	worderCount int
	stated      time.Time
	verbose     bool
	exit        bool
)

func init() {
	stated = time.Now()

}
func Exit() {
	exit = true
}
func Listen(updates chan builder.Update, i int, v bool) {
	worderCount = i
	verbose = v
	statuses = make(map[int]builder.Update)
	for k := 0; k < i; k++ {
		statuses[k] = builder.Update{
			TimeStamp: time.Now(),
			Status:    builder.Pending,
		}
	}
	for {
		select {
		case u := <-updates:
			statuses[u.Worker] = u
		}
	}

}

func failMessage(s string) {
	termPrintln(fmt.Sprintf("[ %s ] %s\n", color.RedString("FAIL"), s))

}

func Run(done chan bool) {

	if !verbose {
		tm.Clear() // Clear current screen
	}

	failed := false
	var failedUpdate builder.Update

	for {
		time.Sleep(time.Millisecond)
		if !verbose {
			tm.MoveCursor(1, 1)
			header := fmt.Sprintf("Building (%s)",
				nstime.NsReadable(time.Since(stated).Nanoseconds()),
			)
			termPrintln(header)
		}

		for worker, update := range statuses {
			if !verbose {
				tm.MoveCursor(worker+2, 1)
			}

			switch update.Status {
			case builder.Pending:
				termPrintln("[ IDLE ]")
			case builder.Started:
				ts := time.Since(update.TimeStamp)
				pbr := ">"

				s := fmt.Sprintf("%s %s (%s)",
					pbr,
					update.Target,
					nstime.NsReadable(ts.Nanoseconds()),
				)
				termPrintln(s)
			case builder.Fail:
				termPrintln("[ IDLE ]")
				exit = true
				failed = true
				failedUpdate = update
			case builder.Success:
				termPrintln("[ IDLE ]")
				break
			}

		}
		if !verbose {
			tm.Flush() // Call it every time at the end of rendering
		}
		if exit {
			if failed {
				if !verbose {
					tm.MoveCursor(worderCount+2, 1)
					failMessage(failedUpdate.Target)
					tm.Flush()
				} else {
					failMessage(failedUpdate.Target)
				}
			}
			done <- true
		}
	}
}
func termPrintln(s string) {
	if verbose {
		//		log.Println(s)
		return
	}
	t := tm.Width()
	line := make([]byte, t)
	for i := 0; i < t; i++ {
		line[i] = []byte(" ")[0]
	}
	line = append([]byte(s), line[len(s):]...)
	tm.Printf(string(line))
}
