package utils

import "testing"

func Test_Utils(t *testing.T) {
	res := UrlEncodeBytesPtrOrNil(nil)
	if res != nil {
		t.Error("UrlEncodeBytesPtrOrNil was not nil")
	}

	zeros := isZeros([]byte(""))
	if !zeros {
		t.Error("isZeros was not true")
	}
	val := 5
	str := StrOrNA(&val)
	if str != "5" {
		t.Error("StrOrNA was not 5")
	}

}
