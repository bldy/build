package skylark

import (
	"bldy.build/build/executor"
	"github.com/google/skylark"
)

type skylarkRule struct {
	name string
	deps []string

	skyFuncLabel string
	skyThread    *skylark.Thread
	args         skylark.Tuple
	kwargs       []skylark.Tuple
	skyFunc      *skylark.Function
	attrs        *skylark.Dict
}

func (s *skylarkRule) Hash() []byte {
	return nil
}
func (s *skylarkRule) GetName() string {
	return s.name
}

func (s *skylarkRule) GetDependencies() []string {
	return s.deps
}

func (s *skylarkRule) Build(e *executor.Executor) error {
	thread := &skylark.Thread{}
	_, err := s.skyFunc.Call(thread, []skylark.Value{}, nil)
	return err
}

func (s *skylarkRule) Installs() map[string]string {
	return nil
}
