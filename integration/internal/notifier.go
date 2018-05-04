package internal

import (
	"log"
	"testing"
	"time"

	"bldy.build/build"
	"bldy.build/build/builder"
	"bldy.build/build/graph"
)

type TestNotifier struct {
	t      *testing.T
	expect *expectNotifier
}

type expectNotifier struct {
	updates map[string]build.Status
	err     error
	done    bool
}

func (e *expectNotifier) Notify(status build.Status, node *graph.Node) {
	log.Printf(node.Label.String())

	e.updates[node.Label.String()] = status
}

func (e *expectNotifier) Failed(err error) {
	e.err = err
}

func (e *expectNotifier) Done(d time.Duration) {
	e.done = true
}

func NewNotifier(t *testing.T) builder.Notifier {
	return &TestNotifier{t, &expectNotifier{
		updates: make(map[string]build.Status),
	}}
}

func (n *TestNotifier) EXPECT() interface{} {
	return n.expect
}

func (n *TestNotifier) Notify(status build.Status, node *graph.Node) {
	if expected, ok := n.expect.updates[node.Label.String()]; !ok {
		log.Printf("expected %d got %d for %s", expected, status, node.Label.String())
		n.t.Fail()
	}
}

func (n *TestNotifier) Failed(err error) {
	n.t.Fail()
	n.t.Log(err)
}

func (n *TestNotifier) Done(d time.Duration) {
	n.t.Logf("finished in %s\n", d)
}
