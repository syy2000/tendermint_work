package scoin

func Filt(s []byte) []byte {

	var t = make([]byte, len(s))
	for i, b := range s {
		if b == '"' {
			t[i] = '$'
		} else {
			t[i] = b
		}
	}
	return t
}

func UnFilt(s []byte) []byte {

	var t = make([]byte, len(s))
	for i, b := range s {
		if b == '$' {
			t[i] = '"'
		} else {
			t[i] = b
		}
	}
	return t

}
