package skylark

import "bldy.build/build/executor"

// Run represents a ctx.actions.run functions in bazel land.
//
// https://docs.bazel.build/versions/master/skylark/lib/actions.html#run
type doNothing struct {
	inputs   []string // List of the output files of the action.
	Mnemonic string   // A one-word description of the action, for example, CppCompile or GoLink.
}

func (r *doNothing) Do(*executor.Executor) error { return nil }
