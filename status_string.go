// Code generated by "stringer -type=Status"; DO NOT EDIT.

package build

import "strconv"

const _Status_name = "SuccessFailPendingStartedFatalWarningBuilding"

var _Status_index = [...]uint8{0, 7, 11, 18, 25, 30, 37, 45}

func (i Status) String() string {
	if i < 0 || i >= Status(len(_Status_index)-1) {
		return "Status(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Status_name[_Status_index[i]:_Status_index[i+1]]
}
