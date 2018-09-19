package skylark

import (
	"errors"

	"bldy.build/build/project"
	"github.com/google/skylark"
)

func _env(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	if s, ok := args.Index(0).(skylark.String); ok {
		env := skylark.String(project.Getenv(string(s)))
		return env, nil
	}
	return nil, errors.New("env only accepts strings as it's first argument")
}
