package skylark

import (
	"errors"
	"fmt"
	"strings"

	"bldy.build/build/skylark/skylarkutils"

	"github.com/google/skylark"
	"github.com/google/skylark/skylarkstruct"
)

var (
	errFormatNotString = errors.New(("formats have to be strings"))
	errMissingAttrs    = errors.New(("attrs are missing"))
)

func format(s skylark.String, attrs *skylarkstruct.Struct) (string, error) {
	if attrs == nil {
		return "", errMissingAttrs
	}
	if input, ok := skylark.AsString(s); ok {
	FORMAT:
		start := strings.IndexAny(input, "%{")
		end := strings.IndexAny(input, "}")
		if start < 0 {
			return input, nil
		}
		advance := 0
		switch input[start] {
		case '{':
			advance = 1
		case '%':
			advance = 2
		}
		keyword := input[start+advance : end]
		key, err := attrs.Attr(keyword)
		if err != nil {
			return "", err
		}
		val, err := skylarkutils.ValueToGo(key)
		if err != nil {
			return "", err
		}
		input = fmt.Sprintf("%s%v%s", input[:start], val, input[end+1:])
		goto FORMAT
	}
	return "", errFormatNotString
}