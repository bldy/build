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
		var d string
		d, s.data = s.data[l-1], s.data[:l-1]
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
	x := pkgStck.peek()

	return x
}

func pushPkg(thread *skylark.Thread, x string) {
	pkgStck := thread.Local(threadKeyPackage).(*pkgStack)

	pkgStck.push(x)
}
func popPkg(thread *skylark.Thread) string {
	pkgStck := thread.Local(threadKeyPackage).(*pkgStack)
	x := pkgStck.pop()

	return x
}

func initPkgStack(thread *skylark.Thread) {
	thread.SetLocal(threadKeyPackage, &pkgStack{})

}