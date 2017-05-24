package harvey

import (
	"crypto/sha1"
	"fmt"
	"io"

	"io/ioutil"

	"bldy.build/build"
)

type DataToC struct {
	Name         string   `data_to_c:"name"`
	Dependencies []string `data_to_c:"deps"`
	Bin          string   `data_to_c:"bin"`
	Prefix       string   `data_to_c:"prefix"`
}

func (dtc *DataToC) GetName() string {
	return dtc.Name
}

func (dtc *DataToC) GetDependencies() []string {
	return dtc.Dependencies
}

func (dtc *DataToC) Hash() []byte {
	h := sha1.New()

	io.WriteString(h, dtc.Name)
	io.WriteString(h, dtc.Bin)
	io.WriteString(h, dtc.Prefix)
	return []byte{}
}

func (dtc *DataToC) Build(e *build.Executor) error {

	inFile, err := e.Open(dtc.Bin)
	if err != nil {
		return err
	}

	in, err := ioutil.ReadAll(inFile)
	if err != nil {
		return err
	}

	total := len(in)
	out, err := e.Create(fmt.Sprintf("%s.c", dtc.Name))

	if err != nil {
		return err
	}

	fmt.Fprintf(out, "unsigned char %vcode[] = {\n", dtc.Prefix)
	for len(in) > 0 {
		for j := 0; j < 16 && len(in) > 0; j++ {
			fmt.Fprintf(out, "0x%02x, ", in[0])
			in = in[1:]
		}
		fmt.Fprintf(out, "\n")

	}

	fmt.Fprintf(out, "0,\n};\nint %vlen = %v;\n", dtc.Prefix, total)
	return nil
}
func (dtc *DataToC) Installs() map[string]string {
	installs := make(map[string]string)
	fname := fmt.Sprintf("%s.c", dtc.Name)
	installs[fname] = fname
	return installs
}
