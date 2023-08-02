package statustable

import "fmt"

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
