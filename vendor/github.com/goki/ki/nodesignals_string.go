// Code generated by "stringer -type=NodeSignals"; DO NOT EDIT.

package ki

import (
	"fmt"
	"strconv"
)

const _NodeSignals_name = "NodeSignalNilNodeSignalUpdatedNodeSignalDeletingNodeSignalDestroyingNodeSignalsN"

var _NodeSignals_index = [...]uint8{0, 13, 30, 48, 68, 80}

func (i NodeSignals) String() string {
	if i < 0 || i >= NodeSignals(len(_NodeSignals_index)-1) {
		return "NodeSignals(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _NodeSignals_name[_NodeSignals_index[i]:_NodeSignals_index[i+1]]
}

func (i *NodeSignals) FromString(s string) error {
	for j := 0; j < len(_NodeSignals_index)-1; j++ {
		if s == _NodeSignals_name[_NodeSignals_index[j]:_NodeSignals_index[j+1]] {
			*i = NodeSignals(j)
			return nil
		}
	}
	return fmt.Errorf("String %v is not a valid option for type NodeSignals", s)
}
