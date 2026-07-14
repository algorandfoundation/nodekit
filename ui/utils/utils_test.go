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

func Test_ShortAddress(t *testing.T) {
	if got := ShortAddress("ALGO123456789"); got != "ALGO..6789" {
		t.Errorf("ShortAddress = %q, want %q", got, "ALGO..6789")
	}
	// Addresses shorter than nine characters are returned unchanged.
	if got := ShortAddress("ABC"); got != "ABC" {
		t.Errorf("ShortAddress = %q, want %q", got, "ABC")
	}
}
