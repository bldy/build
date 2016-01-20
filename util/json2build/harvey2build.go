package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sevki.org/build/util"

	"reflect"

	"github.com/fatih/structs"
)

func main() {
	wd, _ := os.Getwd()
	fs, _ := filepath.Glob(filepath.Join(wd, "*.json"))
	loads := make(map[string]string)

	var clibs []CLib

	for _, x := range fs {
		var ms map[string]CLib

		f, _ := os.Open(x)
		decoder := json.NewDecoder(f)

		decoder.Decode(&ms)

		for name, target := range ms {

			target.Name = name
			for _, inc := range target.Load {
				loads[inc] = "cc_library"
			}

			clibs = append(clibs, target)
		}

	}

	for libtype, _ := range loads {

		if strings.Contains(libtype, "klib") {
			someth := `load("//sys/src/FLAGS", "KLIB_COMPILER_FLAGS")
`
			fmt.Printf(someth)
		} else {
			someth := `load("//sys/src/FLAGS", "LIB_COMPILER_FLAGS")
`
			fmt.Printf(someth)
		}
	}

	fmt.Println("")

	for _, clib := range clibs {
		fmt.Println(clib.String())
	}
}

func (c CLib) String() string {
	kv := make(Strm)
	for k, v := range structs.Map(c) {
		if k == "Load" {
			continue
		}
		if k == "CompilerFlags" {
			continue
		}
		switch v.(type) {
		case string:

		case Dependencies:
			kv[k] = v.(Dependencies).toSSlice()

		default:
			kv[k] = v.(SSlice)
		}

	}

	t := `cc_binary(
	name = "%s",
        copts = %s,
        includes=[
            "//sys/include",
            "//amd64/include",
        ],
	%s
)`
	if strings.Contains(c.Name, "Kernel") {
		fmt.Printf(t, strings.ToLower(strings.Replace(c.Name, "Kernel", "k", -1)), "KLIB_COMPILER_FLAGS", kv)
	} else {
		fmt.Printf(t, strings.ToLower(strings.Replace(c.Name, "Kernel", "k", -1)), "LIB_COMPILER_FLAGS", kv)
	}
	return ""
}

type Strm map[string]SSlice
type SSlice []string
type Dependencies []string

func (s Strm) String() string {
	var strs []string
	for k, v := range s {
		if len(v) == 0 {
			continue
		}
		field, _ := reflect.TypeOf(CLib{}).FieldByName(k)
		strs = append(strs, fmt.Sprintf("%s = %s", field.Tag.Get("cc_library"), v))
	}
	return fmt.Sprintf(strings.Join(strs, ",\n\t")) + ","
}

type CLib struct {
	Name          string       `json:"Name" cc_library:"name"`
	Sources       SSlice       `json:"SourceFiles,omitempty" cc_library:"srcs"`
	Dependencies  Dependencies `json:"Projects,omitempty" cc_library:"deps"`
	Includes      SSlice       `cc_library:"includes"`
	Headers       SSlice       `cc_library:"hdrs" build:"path"`
	CompilerFlags SSlice       `json:"Cflags,omitempty"  cc_library:"copts"`
	LinkerFlags   SSlice       `json:"Oflags,omitempty" cc_library:"linkopts"`
	Load          []string     `json:"Include"`
}

func (c SSlice) String() string {
	t := "[\n\t\t%s,\n\t]"
	for i, s := range c {
		c[i] = fmt.Sprintf("\"%s\"", strings.TrimLeft(s, "-"))
	}
	return fmt.Sprintf(t, strings.Join(c, ",\n\t\t"))
}
func (c Dependencies) toSSlice() SSlice {
	ProjectPath := util.GetProjectPath()
	var t SSlice
	for _, s := range c {
		dir, file := filepath.Split(string(s))
		dir = strings.TrimRight(dir, "/")
		targ := file[:len(file)-len(filepath.Ext(file))]
		if len(dir) > 0 {
			if dir[0] != "/"[0] {
				rel, _ := filepath.Rel(ProjectPath, dir)
				t = append(t, fmt.Sprintf("//%s/%s:%s", rel, dir, targ))
			} else {
				t = append(t, fmt.Sprintf("/%s:%s", dir, targ))
			}

		} else {
			t = append(t, fmt.Sprintf("%s", targ))
		}
	}
	return t
}
func (c Dependencies) String() string {
	t := "[\n\t\t%s\n\t]"

	for i, s := range c {
		dir, file := filepath.Split(string(s))

		dir = strings.TrimRight(dir, "/")
		targ := file[:len(file)-len(filepath.Ext(file))]
		if len(dir) > 0 {

			c[i] = fmt.Sprintf("\"//%s:%s\"", dir, targ)

		} else {
			c[i] = fmt.Sprintf("\"%s\"", targ)
		}
	}
	return fmt.Sprintf(t, strings.Join(c, ",\n\t\t"))
}
