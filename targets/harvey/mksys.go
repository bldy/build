/*
 * This file is part of the UCB release of Plan 9. It is subject to the license
 * terms in the LICENSE file found in the top-level directory of this
 * distribution and at http://akaros.cs.berkeley.edu/files/Plan9License. No
 * part of the UCB release of Plan 9, including this file, may be copied,
 * modified, propagated, or distributed except according to the terms contained
 * in the LICENSE file.
 */

package harvey

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"text/template"

	"sevki.org/build"
	"sevki.org/build/util"
	"sevki.org/lib/prettyprint"
)

type Syscall struct {
	Ret     []string
	Args    []string
	Name    string
	Id      uint32
	Define  string
	Sysname string
	Libname string
	Fudge   string   `json:"-"`
	GoArgs  []string `json:"-"`
	Ret0    string   `json:"-"`
}

type Syserror struct {
	Name   string
	String string
	Id     uint32
}

type Bootmethods struct {
	Name    string
	Config  string
	Connect string
	Arg     string
}

type Sysconf struct {
	Syscalls    []Syscall
	Syserrors   []Syserror
	Bootmethods []Bootmethods
}

type MkSys struct {
	Name         string   `mk_sys:"name"`
	Mode         string   `mk_sys:"mode"`
	ARCH         string   `mk_sys:"arch"`
	SysConf      string   `mk_sys:"sysconf" build:"path"`
	Dependencies []string `mk_sys:"deps"`
	buf          bytes.Buffer
	Source       string
}

func (cl *MkSys) Hash() []byte {

	h := sha1.New()
	io.WriteString(h, cl.Mode)
	io.WriteString(h, cl.ARCH)
	io.WriteString(h, cl.Mode)
	util.HashFiles(h, []string{cl.SysConf})

	return h.Sum(nil)
}

func (mkSys *MkSys) readSysconf() (*Sysconf, error) {
	buf, err := ioutil.ReadFile(mkSys.SysConf)
	if err != nil {
		return nil, fmt.Errorf("os readAll %s:%s", mkSys.SysConf, err.Error())
	}

	var sysconf Sysconf
	err = json.Unmarshal(buf, &sysconf)
	if err != nil {
		return nil, err
	}
	return &sysconf, nil

}
func (mkSys *MkSys) Installs() map[string]string {
	installs := make(map[string]string)
	switch mkSys.Mode {
	case "sys.h", "error.h", "sysdecl.h", "errstr.h":
		installs[filepath.Join("include", mkSys.Mode)] = mkSys.Mode
	case "syscallfiles":
		sysconf, err := mkSys.readSysconf()
		if err != nil {
			return nil
		}
		syscalls := sysconf.Syscalls
		for i := range syscalls {
			if syscalls[i].Define == "" {
				syscalls[i].Define = strings.ToUpper(syscalls[i].Name)
			}
			if syscalls[i].Sysname == "" {
				syscalls[i].Sysname = "sys" + syscalls[i].Name
			}
			if syscalls[i].Libname == "" {
				syscalls[i].Libname = syscalls[i].Name
			}
		}
		for i := range syscalls {
			installs[filepath.Join("9syscall", syscalls[i].Libname+".s")] = syscalls[i].Libname + ".s"
		}
	default:
		installs[mkSys.Mode] = mkSys.Mode
	}
	return installs
}

func (mkSys *MkSys) GetSource() string {
	return mkSys.Source
}

func (mkSys *MkSys) GetName() string {
	return mkSys.Name
}

func (mkSys *MkSys) GetDependencies() []string {
	return mkSys.Dependencies
}

func (mkSys *MkSys) Build(c *build.Context) error {

	c.Println(prettyprint.AsJSON(mkSys))

	sysconf, err := mkSys.readSysconf()
	if err != nil {
		return err
	}

	syscalls := sysconf.Syscalls
	syserrors := sysconf.Syserrors
	bootmethods := sysconf.Bootmethods
	for i := range syscalls {
		if syscalls[i].Define == "" {
			syscalls[i].Define = strings.ToUpper(syscalls[i].Name)
		}
		if syscalls[i].Sysname == "" {
			syscalls[i].Sysname = "sys" + syscalls[i].Name
		}
		if syscalls[i].Libname == "" {
			syscalls[i].Libname = syscalls[i].Name
		}
	}

	outfile, err := c.Create(mkSys.Mode)
	if err != nil {
		return err
	}
	switch mkSys.Mode {
	case "sys_harvey.s":
		if mkSys.ARCH != "amd64" {
			c.Println("ARCH unsupported or not set")
		}
		syscallargs := []string{"DI", "SI", "DX", "R10", "R8", "R9"}
		//funcallregs := []string{ "DI", "SI", "DX", "CX", "R8", "R9" };
		for i := range syscalls {
			goargs := []string{}
			fpoff := 0
			for k := range syscalls[i].Args {
				switch syscalls[i].Args[k] {
				case "int32_t", "uint32_t":
					goargs = append(goargs, fmt.Sprintf("MOVL	arg%d+%d(FP), %s", k, fpoff, syscallargs[k]))
					fpoff += 4
				case "void*", "char*", "char**", "uint8_t*", "int32_t*", "uint64_t*", "int64_t*", "int64_t":
					fpoff = (fpoff + 7) & ^7
					goargs = append(goargs, fmt.Sprintf("MOVQ	arg%d+%d(FP), %s", k, fpoff, syscallargs[k]))
					fpoff += 8
				default:
					fmt.Errorf("unsupported arg %s in syscall: %v", syscalls[i].Args[k], syscalls[i])
				}
			}
			syscalls[i].GoArgs = goargs
			switch syscalls[i].Ret[0] {
			case "int32_t", "uint32_t":
				syscalls[i].Ret0 = fmt.Sprintf("MOVL	AX, ret+%d(FP)", fpoff)
				fpoff += 4
			case "void*", "char*", "char**", "uint8_t*", "int32_t*", "uint64_t*", "int64_t*", "int64_t":
				fpoff = (fpoff + 7) & ^7
				syscalls[i].Ret0 = fmt.Sprintf("MOVQ	AX, ret+%d(FP)", fpoff)
				fpoff += 8
			default:
				return fmt.Errorf("unsupported Ret[0] in syscall: %v", syscalls[i])
			}
		}
		tmpl, err := template.New("sys_harvey.s").Parse(`/* automatically generated by mksys */
/* System calls for AMD64, Harvey */
#include "go_asm.h"
#include "go_tls.h"
#include "textflag.h"
{{ range . }}
TEXT runtime·{{ .Libname }}(SB),NOSPLIT,$0
{{ range .GoArgs }}	{{ . }}
{{ end }}	MOVQ	${{ .Id }}, AX
	SYSCALL
	{{ .Ret0 }}
	RET
{{ end }}
`)
		if err != nil {
			return err
		}
		err = tmpl.Execute(outfile, syscalls)
		if err != nil {
			return err
		}

	case "syscallfiles":
		if mkSys.ARCH != "amd64" {
			return fmt.Errorf("ARCH unsupported or not set")
		}
		tmpl, err := template.New("syscall.s").Parse(`/* automatically generated by mksys */
.globl	{{ .Libname }}
{{ .Libname }}:
	movq %rcx, %r10 /* rcx gets smashed by systenter. Use r10.*/
	movq ${{ .Id }},%rax  /* Put the system call into rax, just like linux. */
	syscall
	ret
`)
		if err != nil {
			return err
		}

		for i := range syscalls {

			file, err := c.Create(syscalls[i].Libname + ".s")
			if err != nil {
				return err
			}

			err = tmpl.Execute(file, syscalls[i])
			if err != nil {
				return err
			}

			err = file.Close()
			if err != nil {
				return err
			}
		}
	case "sysnum.go":
		tmpl, err := template.New("sysnum.go").Parse(`// automatically generated by mksys
package syscall
const(
{{ range . }}	SYS_{{ .Define }} = {{ .Id }}
{{ end }}
)
`)
		err = tmpl.Execute(outfile, syscalls)
		if err != nil {
			return err
		}

	case "sys.h":
		tmpl, err := template.New("sys.h").Parse(`/* automatically generated by mksys */
{{ range . }}#define {{ .Define }} {{ .Id }}
{{ end }}
`)
		err = tmpl.Execute(outfile, syscalls)
		if err != nil {
			return err
		}

	case "sysdecl.h":
		tmpl, err := template.New("sysdecl.h").Parse(`/* automatically generated by mksys */
{{ range . }}extern {{ .Ret0 }} {{ .Libname }}({{ range $i, $e := .Args }}{{ if $i }}, {{ end }}{{ $e }}{{ end }});
{{ end }}
`)
		err = tmpl.Execute(outfile, syscalls)
		if err != nil {
			return err
		}

	case "systab.c":
		for i := range syscalls {
			var fudge string
			switch syscalls[i].Ret[0] {
			case "int32_t":
				fudge = "{ .i = -1 }"
			case "int64_t":
				fudge = "{ .vl = -1ll }"
			case "void*", "char*":
				fudge = "{ .v = (void*)-1ll }"
			default:
				return fmt.Errorf("unsupported Ret[0] in syscall: %v", syscalls[i])
			}
			if syscalls[i].Fudge == "" {
				syscalls[i].Fudge = fudge
			}

			syscalls[i].Ret0 = syscalls[i].Ret[0]
		}
		tmpl, err := template.New("systab.c").Parse(`/* automatically generated by mksys */
#include "u.h"
#include "../port/lib.h"
#include "mem.h"
#include "dat.h"
#include "fns.h"
#include "../../libc/9syscall/sys.h"
{{ range . }}extern void {{ .Sysname }}(Ar0*, ...);
{{ end }}
Systab systab[] = {
{{ range . }}[{{ .Define }}] { "{{ .Name }}", {{ .Sysname }}, {{ .Fudge }} },
{{ end }}
};
int nsyscall = nelem(systab);
`)
		err = tmpl.Execute(outfile, syscalls)
		if err != nil {
			return err
		}

	case "error.h":
		tmpl, err := template.New("error.h").Parse(`/* automatically generated by mksys */
{{ range . }}extern char {{ .Name }}[]; /* {{ .String }} */
{{ end }}
`)
		err = tmpl.Execute(outfile, syserrors)
		if err != nil {
			return err
		}

	case "errstr.h":
		tmpl, err := template.New("errstr.h").Parse(`/* automatically generated by mksys */
{{ range . }}char {{ .Name }}[] = "{{ .String }}";
{{ end }}
`)
		err = tmpl.Execute(outfile, syserrors)
		if err != nil {
			return err
		}
	case "bootamd64cpu.c":
		tmpl, err := template.New("bootamd64cpu.c").Parse(`/* automatically generated by mksys */
#include <u.h>
#include <libc.h>
#include <boot.h>
Method method[] = {
{{ range . }}{ "{{.Name}}", {{.Config}}, {{.Connect}}, "{{.Arg}}", },
{{ end }}
	{ nil },
};
int cpuflag = 1;
char* rootdir = "/root";
char* bootdisk = "#S/sdE0/";
extern void boot(int, char**);
void
main(int argc, char **argv)
{
		boot(argc, argv);
}
int (*cfs)(int) = 0;
`)
		err = tmpl.Execute(outfile, bootmethods)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}
