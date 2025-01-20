package utils

import (
	"encoding/base64"
	"fmt"
)

func toPtr[T any](constVar T) *T { return &constVar }

func Base64EncodeBytesPtrOrNil(b []byte) *string {
	if b == nil || len(b) == 0 || isZeros(b) {
		return nil
	}
	return toPtr(base64.StdEncoding.EncodeToString(b))
}

func UrlEncodeBytesPtrOrNil(b []byte) *string {
	if b == nil || len(b) == 0 || isZeros(b) {
		return nil
	}
	return toPtr(base64.RawURLEncoding.EncodeToString(b))
}

func isZeros(b []byte) bool {
	for i := 0; i < len(b); i++ {
		if b[i] != 0 {
			return false
		}
	}
	return true
}

func StrOrNA(value *int) string {
	if value == nil {
		return "N/A"
	}
	return IntToStr(*value)
}
func IntToStr(number int) string {
	return fmt.Sprintf("%d", number)
}
