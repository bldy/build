package harvey

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"

	"os"
	"text/template"

	"io"
	"strings"

	"bldy.build/build/racy"

	"path/filepath"

	"bldy.build/build"
)

type Config struct {
	Name         string            `config:"name"`
	Dependencies []string          `config:"deps"`
	RamFiles     []string          `config:"ramfiles" build:"path"`
	Code         []string          `config:"code"`
	Dev          []string          `config:"dev"`
	Ip           []string          `config:"ip"`
	Link         []string          `config:"link"`
	Sd           []string          `config:"sd"`
	Uart         []string          `config:"uart"`
	VGA          []string          `config:"vga"`
	Bins         map[string]string `config:"bins"`
}

func split(s string, c string) string {
	i := strings.Index(s, c)
	if i < 0 {
		return s
	}

	return s[i+1:]
}
func (k *Config) Hash() []byte {

	h := sha1.New()
	racy.HashFiles(h, k.RamFiles)
	racy.HashStrings(h, k.Code)
	racy.HashStrings(h, k.Dev)
	racy.HashStrings(h, k.Ip)
	racy.HashStrings(h, k.Link)
	racy.HashStrings(h, k.Sd)
	racy.HashStrings(h, k.Uart)
	racy.HashStrings(h, k.VGA)
	io.WriteString(h, k.Name)
	return h.Sum(nil)
}

func (k *Config) Build(c *build.Runner) error {

	var rootcodes []string
	var rootnames []string
	for _, dep := range k.Dependencies {
		name := split(dep, ":")
		path := filepath.Join("bin", name)
		for k, v := range k.Bins {
			if k == name {
				path = filepath.Join("bin", v)
				break
			}
		}
		code, err := data2c(name, path, c)

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
		Config    Config
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

func (k *Config) Installs() map[string]string {
	exports := make(map[string]string)
	path := fmt.Sprintf("%s.c", k.Name)
	exports[path] = path

	return exports
}

func (k *Config) GetName() string {
	return k.Name
}
func (k *Config) GetDependencies() []string {
	return k.Dependencies
}

// data2c takes the file at path and creates a C byte array containing it.
func data2c(name string, path string, c *build.Runner) (string, error) {
	var out []byte
	var in []byte
	var file *os.File
	var err error
	if file, err = c.Open(path); err != nil {
		return "", fmt.Errorf("open couldn't find %s: %s", path, err.Error())

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
