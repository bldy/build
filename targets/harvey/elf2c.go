package harvey

import (
	"crypto/sha1"
	"debug/elf"
	"fmt"
	"io"
	"math"
	"path"
	"path/filepath"

	"sevki.org/build"
)

type ElfToC struct {
	Name         string   `elf_to_c:"name"`
	Dependencies []string `elf_to_c:"deps"`
	Elf          string   `elf_to_c:"elf"`
}

func (etc *ElfToC) GetName() string {
	return etc.Name
}

func (etc *ElfToC) GetDependencies() []string {
	return etc.Dependencies
}

func (etc *ElfToC) Installs() map[string]string {
	i := make(map[string]string)

	i[filepath.Join("include", fmt.Sprintf("%s.h", etc.Name))] = fmt.Sprintf("%s.h", etc.Name)
	return i
}

func (etc *ElfToC) Hash() []byte {
	h := sha1.New()

	io.WriteString(h, etc.Name)
	io.WriteString(h, etc.Elf)
	return []byte{}
}

func (etc *ElfToC) Build(c *build.Context) error {
	fileName := ""

	if xf, err := c.Open(etc.Elf); err != nil {
		return fmt.Errorf("open :%s", err.Error())
	} else {
		fileName = xf.Name()
		xf.Close()
	}

	f, err := elf.Open(fileName)
	if err != nil {
		return err
	}
	var dataend, codeend, end uint64
	var datastart, codestart, start uint64
	datastart, codestart, start = math.MaxUint64, math.MaxUint64, math.MaxUint64
	mem := []byte{}
	for _, v := range f.Progs {
		if v.Type != elf.PT_LOAD {
			continue
		}
		c.Printf("processing %v\n", v)
		// MUST alignt to 2M page boundary.
		// then MUST allocate a []byte that
		// is the right size. And MUST
		// see if by some off chance it
		// joins to a pre-existing segment.
		// It's easier than it seems. We produce ONE text
		// array and ONE data array. So it's a matter of creating
		// a virtual memory space with an assumed starting point of
		// 0x200000, and filling it. We just grow that as needed.

		curstart := v.Vaddr & ^uint64(0xfff) // 0x1fffff)
		curend := v.Vaddr + v.Memsz
		c.Printf("s %x e %x\n", curstart, curend)
		if curend > end {
			nmem := make([]byte, curend)
			copy(nmem, mem)
			mem = nmem
		}
		if curstart < start {
			start = curstart
		}

		if v.Flags&elf.PF_X == elf.PF_X {
			if curstart < codestart {
				codestart = curstart
			}
			if curend > codeend {
				codeend = curend
			}
			c.Printf("code s %v e %v\n", codestart, codeend)
		} else {
			if curstart < datastart {
				datastart = curstart
			}
			if curend > dataend {
				dataend = curend
			}
			c.Printf("data s %v e %v\n", datastart, dataend)
		}
		for i := uint64(0); i < v.Filesz; i++ {
			if amt, err := v.ReadAt(mem[v.Vaddr+i:], int64(i)); err != nil && err != io.EOF {
				err := fmt.Errorf("%v: %v\n", amt, err)
				c.Println(err)
				return err
			} else if amt == 0 {
				if i < v.Filesz {
					err := fmt.Errorf("%v: Short read: %v of %v\n", v, i, v.Filesz)
					c.Println(err)
					return err
				}
				break
			} else {
				i = i + uint64(amt)
				c.Println("i now %v\n", i)
			}
		}
		c.Println("Processed %v\n", v)
	}
	c.Printf("gencode\n")
	// Gen code to stdout. For each file, create an array, a start, and an end variable.
	outfile, err := c.Create(fmt.Sprintf("%s.h", etc.Name))
	if err != nil {
		return err
	}

	_, file := path.Split(etc.Elf)

	gencode(outfile, file, "code", mem, codestart, codeend)
	gencode(outfile, file, "data", mem, datastart, dataend)

	return nil
}

func gencode(w io.Writer, n, t string, m []byte, start, end uint64) {

	fmt.Fprintf(w, "int %v_%v_start = %v;\n", n, t, start)
	fmt.Fprintf(w, "int %v_%v_end = %v;\n", n, t, end)
	fmt.Fprintf(w, "int %v_%v_len = %v;\n", n, t, end-start)
	fmt.Fprintf(w, "uint8_t %v_%v_out[] = {\n", n, t)
	for i := uint64(start); i < end; i += 16 {
		for j := uint64(0); i+j < end && j < 16; j++ {
			fmt.Fprintf(w, "0x%02x, ", m[j+i])
		}
		fmt.Fprintf(w, "\n")
	}
	fmt.Fprintf(w, "};\n")
}
