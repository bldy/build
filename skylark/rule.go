package skylark

import (
	"encoding/binary"
	"io"
	"log"

	"bldy.build/build/executor"
	"bldy.build/build/racy"
	"github.com/google/skylark"
)

type skylarkRule struct {
	Name         string
	Dependencies []string

	skyFuncLabel string
	skyThread    *skylark.Thread
	args         skylark.Tuple
	kwargs       []skylark.Tuple
	skyFunc      *skylark.Function
	attrs        *skylark.Dict
}

func (s *skylarkRule) Build(e *executor.Executor) error {
	ctx := skylarkContext{
		label: s.skyFuncLabel,
		attrs: s.attrs,
	}
	thread := &skylark.Thread{}
	_, err := s.skyFunc.Call(thread, []skylark.Value{&ctx}, nil)
	return err
}

func (s *skylarkRule) Hash() []byte {
	h := racy.New()
	io.WriteString(h, s.skyFuncLabel)
	funcHash, err := s.skyFunc.Hash()
	if err != nil {
		log.Fatal(err)
	}
	if err := binary.Write(h, binary.BigEndian, funcHash); err != nil {
		log.Fatal(err)
	}
	return h.Sum(nil)
}

func (s *skylarkRule) GetName() string {
	return s.Name
}

func (s *skylarkRule) GetDependencies() []string {
	return s.Dependencies
}
func (s *skylarkRule) Installs() map[string]string {
	return nil
}
