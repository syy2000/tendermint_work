package statustable

import (
	"bytes"
	"fmt"
	"os"
)

func LonggestCommonPrefix(key1, key2 []byte) int {
	var n int
	if len(key1) < len(key2) {
		n = len(key1)
	} else {
		n = len(key2)
	}
	for i := 0; i < n; i++ {
		if key1[i] != key2[i] {
			return i
		}
	}
	return n
}

func PrefixIndex(u byte) int {
	out := int(int8(u))
	if '0' <= u && u <= '9' {
		return out - int(int8('0'))
	} else if 'a' <= u && u <= 'f' {
		return out - int(int8('a')) + 10
	} else if 'A' <= u && u <= 'F' {
		return out - int(int8('A')) + 10
	} else {
		return -1
	}
}

func Str2Hex(k string) []byte {
	return []byte(fmt.Sprintf("%x", k))
}
func Byte2Hex(k []byte) []byte {
	return []byte(fmt.Sprintf("%x", k))
}

func join2Bytes(a, b []byte) []byte {
	return bytes.Join([][]byte{a, b}, []byte{})
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
