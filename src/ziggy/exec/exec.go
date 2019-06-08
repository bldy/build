package exec

import (
	"fmt"
	"os"

	"bldy.build/bldy/src/build"

	"bldy.build/bldy/src/executor"
	"bldy.build/bldy/src/ziggy/ziggyutils"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type ExecRuntime interface {
	build.Runtime

	Exec(cmd string, env, args []string) error
}

type ActionModule struct {
	starlarkstruct.Module
	calls []executor.Action
}

func (m *ActionModule) record(a executor.Action) {
	m.calls = append(m.calls, a)
}

func (m *ActionModule) newRun(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	i := &run{}
	if err := ziggyutils.UnpackStruct(i, kwargs); err != nil {
		return nil, err
	}

	m.record(i)
	return starlark.None, nil
}

func newRun(kwargs []starlark.Tuple) (executor.Action, error) {
	return nil, nil
}

func New() ActionModule {
	var module = ActionModule{
		Module: starlarkstruct.Module{
			Name:    "exec",
			Members: map[string]starlark.Value{},
		},
	}
	module.Members["run"] = starlark.NewBuiltin("run", module.newRun)
	return module
}

type run struct {
	Outputs               []string          `ziggy:"outputs"`                // List of the output files of the action.
	Files                 []string          `ziggy:"files"`                  // List of the input files of the action.
	Executable            string            `ziggy:"executable"`             // The executable file to be called by the action.
	Arguments             []string          `ziggy:"arguments"`              // Command line arguments of the action. Must be a list of strings or actions.args() objects.
	Mnemonic              string            `ziggy:"mnemonic"`               // A one-word description of the action, for example, CppCompile or GoLink.
	ProgressMessage       string            `ziggy:"progress_message"`       // Progress message to show to the user during the build.
	UseDefaultShellEnv    bool              `ziggy:"use_default_shell"`      // Whether the action should use the built in shell environment or not.
	Env                   map[string]string `ziggy:"env"`                    // Sets the dictionary of environment variables.
	ExecutionRequirements map[string]string `ziggy:"execution_requirements"` // Information for scheduling the action. See tags for useful keys.
}

func (r *run) Do(rt build.Runtime) error {
	runtime, ok := rt.(ExecRuntime)
	if !ok {
		return fmt.Errorf("expected execution runtime got %T instead", runtime)
	}
	env := os.Environ()
	if !r.UseDefaultShellEnv {
		env = []string{}
	}
	for k, v := range r.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	runtime.Printf(r.ProgressMessage)
	return runtime.Exec(r.Executable, env, r.Arguments)
}
