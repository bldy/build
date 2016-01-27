package harvey

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"

	"os"
	"text/template"

	"io"
	"strings"

	"path/filepath"

	"sevki.org/build"
	"sevki.org/lib/prettyprint"
)

type Kernel struct {
	Name         string   `kernel:"name"`
	Dependencies []string `kernel:"deps"`
	RamFiles     []string `kernel:"ramfiles" build:"path"`
	Code         []string `kernel:"code"`
	Dev          []string `kernel:"dev"`
	Ip           []string `kernel:"ip"`
	Link         []string `kernel:"link"`
	Sd           []string `kernel:"sd"`
	Uart         []string `kernel:"uart"`
	VGA          []string `kernel:"vga"`
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

	io.WriteString(h, k.Name)
	io.WriteString(h, prettyprint.AsJSON(k))
	return h.Sum(nil)
}

func (k *Kernel) Build(c *build.Context) error {

	var rootcodes []string
	var rootnames []string
	for _, dep := range k.Dependencies {
		name := split(dep, ":")
		code, err := data2c(name, filepath.Join("bin", name), c)

		if err != nil {
			return err
		}
		rootcodes = append(rootcodes, code)
		rootnames = append(rootnames, name)
	}
	for _, p := range k.RamFiles {
		name := filepath.Base(p)
		code, err := data2c(name, p, c)

		if err != nil {
			return err
		}
		rootcodes = append(rootcodes, code)
		rootnames = append(rootnames, name)
	}
	path := fmt.Sprintf("%s.c", k.Name)
	vars := struct {
		Path      string
		Config    Kernel
		Rootnames []string
		Rootcodes []string
	}{
		path,
		*k,
		rootnames,
		rootcodes,
	}
	tmpl := template.Must(template.New("kernconf").Parse(kernconfTmpl))
	f, err := c.Create(path)
	if err != nil {
		return nil
	}
	return tmpl.Execute(f, vars)

}

func (k *Kernel) Installs() map[string]string {
	exports := make(map[string]string)
	path := fmt.Sprintf("%s.c", k.Name)
	exports[path] =path

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
	var file *os.File
	var err error
	if file, err = c.Open(path); err != nil {
		return "", fmt.Errorf("open :%s", err.Error())

	}

	if in, err = ioutil.ReadAll(file); err != nil {
		return "", nil
	}

	file.Close()

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

type kernconfig struct {
	Code []string
	Dev  []string
	Ip   []string
	Link []string
	Sd   []string
	Uart []string
	VGA  []string
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
