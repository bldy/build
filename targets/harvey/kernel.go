package harvey

import (
	"bytes"
	"crypto/sha1"
	"debug/elf"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"io"
	"strings"

	"path/filepath"

	"sevki.org/build"
	"sevki.org/build/targets/cc"
	"sevki.org/build/util"
	"sevki.org/lib/prettyprint"
)

type Kernel struct {
	Name            string           `kernel:"name"`
	Sources         []string         `kernel:"srcs" build:"path"`
	Dependencies    []string         `kernel:"deps"`
	Includes        cc.Includes      `kernel:"includes" build:"path"`
	Headers         []string         `kernel:"hdrs" build:"path"`
	CompilerOptions cc.CompilerFlags `kernel:"copts"`
	LinkerOptions   []string         `kernel:"linkopts"`
	LinkerFile      string           `kernel:"ld" build:"path"`
	RamFiles        []string         `kernel:"ramfiles"`
	Code            []string
	Dev             []string
	Ip              []string
	Link            []string
	Sd              []string
	Uart            []string
	VGA             []string
}

func split(s string, c string) string {
	i := strings.Index(s, c)
	if i < 0 {
		return s
	}

	return s[i+1:]
}
func (k *Kernel) Hash() []byte {

	h := sha1.New()
	io.WriteString(h, cc.CCVersion)
	io.WriteString(h, k.Name)
	util.HashFiles(h, k.Includes)
	util.HashFiles(h, []string(k.Sources))
	util.HashStrings(h, k.CompilerOptions)
	util.HashStrings(h, k.LinkerOptions)
	return h.Sum(nil)
}

func (k *Kernel) Build(c *build.Context) error {
	c.Println(prettyprint.AsJSON(k))
	params := []string{"-c"}
	params = append(params, k.CompilerOptions...)
	params = append(params, k.Sources...)

	params = append(params, k.Includes.Includes()...)

	if err := c.Exec(cc.Compiler(), cc.CCENV, params); err != nil {
		return fmt.Errorf(err.Error())
	}

	ldparams := []string{"-o", k.Name}
	ldparams = append(ldparams, k.LinkerOptions...)

	if k.LinkerFile != "" {
		ldparams = append(ldparams, k.LinkerFile)
	}

	// This is done under the assumption that each src file put in this thing
	// here will comeout as a .o file
	for _, f := range k.Sources {
		_, fname := filepath.Split(f)
		ldparams = append(ldparams, fmt.Sprintf("%s.o", fname[:strings.LastIndex(fname, ".")]))
	}

	ldparams = append(ldparams, "-L", "lib")

	for _, dep := range k.Dependencies {
		d := split(dep, ":")
		if len(d) < 3 {
			continue
		}
		if d[:3] == "lib" {
			ldparams = append(ldparams, fmt.Sprintf("-l%s", d[3:]))
		}
	}

	if err := c.Exec(cc.Linker(), cc.CCENV, ldparams); err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

func (k *Kernel) Installs() map[string]string {
	exports := make(map[string]string)

	exports[filepath.Join("bin", k.Name)] = k.Name

	return exports
}

func (k *Kernel) GetName() string {
	return k.Name
}
func (k *Kernel) GetDependencies() []string {
	return k.Dependencies
}

// data2c takes the file at path and creates a C byte array containing it.
func data2c(name string, path string, c *build.Context) (string, error) {
	var out []byte
	var in []byte
	fileName := ""
	if xf, err := c.Open(path); err != nil {
		return "", fmt.Errorf("open :%s", err.Error())
	} else {
		fileName = xf.Name()
		xf.Close()
	}

	if elf, err := elf.Open(path); err == nil {
		elf.Close()
		cwd, err := os.Getwd()
		tmpf, err := ioutil.TempFile(cwd, name)
		if err != nil {
			return "", nil
		}

		in, err = ioutil.ReadAll(tmpf)
		if err != nil {
			return "", nil
		}
		tmpf.Close()
		os.Remove(tmpf.Name())
	} else {
		var file *os.File
		var err error

		file, err = os.Open(path)
		if err != nil {
			return "", nil
		}

		in, err = ioutil.ReadAll(file)
		if err != nil {
			return "", nil
		}

		file.Close()
	}

	total := len(in)

	out = []byte(fmt.Sprintf("static unsigned char ramfs_%s_code[] = {\n", name))
	for len(in) > 0 {
		for j := 0; j < 16 && len(in) > 0; j++ {
			out = append(out, []byte(fmt.Sprintf("0x%02x, ", in[0]))...)
			in = in[1:]
		}
		out = append(out, '\n')
	}

	out = append(out, []byte(fmt.Sprintf("0,\n};\nint ramfs_%s_len = %v;\n", name, total))...)

	return string(out), nil
}

// confcode creates a kernel configuration header.
func confcode(path string, k *Kernel, c *build.Context) error {
	var rootcodes []string
	var rootnames []string
	for name, path := range k.RamFiles {
		code, err := data2c(k.Name, filepath.Join("bin", k.Name), c)
		if err != nil {
			return err
		}
		rootcodes = append(rootcodes, code)
		rootnames = append(rootnames, k.Name)
	}

	vars := struct {
		Path      string
		Config    kernconfig
		Rootnames []string
		Rootcodes []string
	}{
		path,
		k.Config,
		rootnames,
		rootcodes,
	}

	tmpl := template.Must(template.New("kernconf").Parse(kernconfTmpl))
	codebuf := &bytes.Buffer{}
	return codebuf.Bytes()
}

const kernconfTmpl = `
#include "u.h"
#include "../port/lib.h"
#include "mem.h"
#include "dat.h"
#include "fns.h"
#include <error.h>
#include "io.h"
void
rdb(void)
{
	splhi();
	iprint("rdb...not installed\n");
	for(;;);
}
{{ range .Rootcodes }}
{{ . }}
{{ end }}
{{ range .Config.Dev }}extern Dev {{ . }}devtab;
{{ end }}
Dev *devtab[] = {
{{ range .Config.Dev }}
	&{{ . }}devtab,
{{ end }}
	nil,
};
{{ range .Config.Link }}extern void {{ . }}link(void);
{{ end }}
void
links(void)
{
{{ range .Rootnames }}addbootfile("{{ . }}", ramfs_{{ . }}_code, ramfs_{{ . }}_len);
{{ end }}
{{ range .Config.Link }}{{ . }}link();
{{ end }}
}
#include "../ip/ip.h"
{{ range .Config.Ip }}extern void {{ . }}init(Fs*);
{{ end }}
void (*ipprotoinit[])(Fs*) = {
{{ range .Config.Ip }}	{{ . }}init,
{{ end }}
	nil,
};
#include "../port/sd.h"
{{ range .Config.Sd }}extern SDifc {{ . }}ifc;
{{ end }}
SDifc* sdifc[] = {
{{ range .Config.Sd }}	&{{ . }}ifc,
{{ end }}
	nil,
};
{{ range .Config.Uart }}extern PhysUart {{ . }}physuart;
{{ end }}
PhysUart* physuart[] = {
{{ range .Config.Uart }}	&{{ . }}physuart,
{{ end }}
	nil,
};
#define	Image	IMAGE
#include <draw.h>
#include <memdraw.h>
#include <cursor.h>
#include "screen.h"
{{ range .Config.VGA }}extern VGAdev {{ . }}dev;
{{ end }}
VGAdev* vgadev[] = {
{{ range .Config.VGA }}	&{{ . }}dev,
{{ end }}
	nil,
};
{{ range .Config.VGA }}extern VGAcur {{ . }}cur;
{{ end }}
VGAcur* vgacur[] = {
{{ range .Config.VGA }}	&{{ . }}cur,
{{ end }}
	nil,
};
Physseg physseg[8] = {
	{
		.attr = SG_SHARED,
		.name = "shared",
		.size = SEGMAXPG,
	},
	{
		.attr = SG_BSS,
		.name = "memory",
		.size = SEGMAXPG,
	},
};
int nphysseg = 8;
{{ range .Config.Code }}{{ . }}
{{ end }}
char* conffile = "{{ .Path }}";
`
