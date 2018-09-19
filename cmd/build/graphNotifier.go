package build

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"

	"bldy.build/build"
	"bldy.build/build/graph"
)

type terminalNotifier struct {
	mut          *sync.Mutex
	workerStatus []build.Status
	w            *tabwriter.Writer
}

func newNotifier(workers int) *terminalNotifier {
	l := make([]build.Status, workers)
	for i, _ := range l {
		fmt.Printf(" %d ", i+1)
		l[i] = build.Pending
	}
	fmt.Println()
	return &terminalNotifier{
		mut:          &sync.Mutex{},
		workerStatus: l,
	}
}

func (t *terminalNotifier) Update(n *graph.Node) {
	t.mut.Lock()
	i, err := strconv.Atoi(n.Worker)
	if err != nil {
		panic(err)
	}
	oldStatus := t.workerStatus[i]
	t.workerStatus[i] = n.Status
	graph := ""
	for j, s := range t.workerStatus {
		graph += " "
		switch s {
		case build.Building:
			if i == j && oldStatus == build.Pending {
				graph += "+"
			} else {
				graph += "|"
			}
		case build.Success:
			graph += "$"
			t.workerStatus[i] = build.Pending
		case build.Fail:
			graph += "X"
			t.workerStatus[i] = build.Pending
		default:
			graph += " "
		}
		graph += " "

	}
	fmt.Printf("%s %s\t(cached = %v\tstatus = %s\tworker = %d)\n", graph, n.Label.String(), n.Cached, n.Status, i+1)

	t.mut.Unlock()
}

func (t *terminalNotifier) Error(err error) {
	log.Fatal(err)
}

func (t *terminalNotifier) Done(d time.Duration) {
	fmt.Printf("finished building in %s\n", d)
}