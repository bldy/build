package skylark

// Run represents a ctx.actions.run functions in bazel land.
//
//
// https://docs.bazel.build/versions/master/skylark/lib/actions.html#run
type run struct {
	Outputs               []string          // List of the output files of the action.
	Files                 []string          // List of the input files of the action.
	Executable            string            // The executable file to be called by the action.
	Arguments             []string          // Command line arguments of the action. Must be a list of strings or actions.args() objects.
	Mnemonic              string            // A one-word description of the action, for example, CppCompile or GoLink.
	ProgressMessage       string            // Progress message to show to the user during the build, for example, "Compiling foo.cc to create foo.o".
	UseDefaultShellEnv    bool              // Whether the action should use the built in shell environment or not.
	Env                   map[string]string // Sets the dictionary of environment variables.
	ExecutionRequirements map[string]string // Information for scheduling the action. See tags for useful keys.
}

func (r *run) Do() error {
	return nil
}
