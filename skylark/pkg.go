package skylark

import (
	"github.com/google/skylark"
)

type pkgStack struct {
	data []string
}

func (s *pkgStack) push(v string) {
	s.data = append(s.data, v)
}

func (s *pkgStack) pop() string {
	l := len(s.data)
	if l > 1 {
		d := s.data[l-1]
		return d
	} else {
		panic("this should never Ever ever EVAAA happn")
	}
}
func (s *pkgStack) peek() string {
	l := len(s.data)
	if l > 0 {
		return s.data[l-1]
	} else {
		panic("this should never Ever ever EVAAA happn")
	}
}

func getPkg(thread *skylark.Thread) string {
	pkgStck := thread.Local(threadKeyPackage).(*pkgStack)
	return pkgStck.peek()
}

func pushPkg(thread *skylark.Thread, v string) {
	pkgStck := thread.Local(threadKeyPackage).(*pkgStack)
	pkgStck.push(v)
}
func popPkg(thread *skylark.Thread) string {
	pkgStck := thread.Local(threadKeyPackage).(*pkgStack)
	return pkgStck.pop()
}

func initPkgStack(thread *skylark.Thread) {
	thread.SetLocal(threadKeyPackage, &pkgStack{})

}