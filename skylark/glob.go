package skylark

import (
	"log"
	"path/filepath"
	"regexp"

	"bldy.build/build/label"
	"bldy.build/build/workspace"
	"github.com/google/skylark"
	"github.com/pkg/errors"
)

var rootLabel = label.Label("//.")

func (s *skylarkVM) glob(thread *skylark.Thread, fn *skylark.Builtin, args skylark.Tuple, kwargs []skylark.Tuple) (skylark.Value, error) {
	pkg := getPkg(thread)
	excl := new(skylark.List)

	var excludeDirectories skylark.Int

	_ = pkg

	err := skylark.UnpackArgs(fn.Name(), skylark.Tuple{}, kwargs, "excludes?", &excl, "exclude_directories?", &excludeDirectories)
	if err != nil {
		log.Println("glob.unpack", err)
	}
	includes := []string{}
	excludes := []string{}
	files := []string{}

	if excl != nil {
		i := excl.Iterate()
		var p skylark.Value
		for i.Next(&p) {
			if pattern, ok := skylark.AsString(p); ok {
				excludes = append(excludes, pattern)
			}
		}
	}
	{
		i := args.Iterate()
		var p skylark.Value
		for i.Next(&p) {
			if pattern, ok := skylark.AsString(p); ok {
				x := absolutePath(s.ws, pkg, pattern)
				includes = append(includes, x)

				if matches, err := filepath.Glob(x); err != nil {
					return nil, err
				} else {
					files = append(files, matches...)
				}
			}
		}
	}
	x := []skylark.Value{}

	wsPath := s.ws.AbsPath()

	// filter out exclusions
	for _, excl := range excludes {
		for i := 0; i < len(files); i++ {
			f := files[i]
			matched, err := regexp.MatchString(excl, f)
			if err != nil {
				return nil, errors.Wrap(err, "glob")
			}
			if matched {
				files = append(files[:i], files[i+1:]...)
			}
		}
	}

	for _, f := range files {
		rel, _ := filepath.Rel(wsPath, f)
		//		x = append(x, file.New(label.Label(rel), rootLabel, s.ws))
		x = append(x, skylark.String(rel))
	}

	return skylark.NewList(x), nil

}

func absolutePath(w workspace.Workspace, pkg string, pattern string) string {
	return filepath.Join(w.AbsPath(), pkg, pattern)
}
