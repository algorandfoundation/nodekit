package utils

import (
	"encoding/base64"
	"fmt"
	"github.com/charmbracelet/log"
	"strconv"
)

func toPtr[T any](constVar T) *T { return &constVar }

// ShortAddress abbreviates an Algorand address to its first and last four
// characters joined by two dots (e.g. "ABCD..WXYZ"). Addresses shorter than
// nine characters are returned unchanged.
func ShortAddress(address string) string {
	if len(address) < 9 {
		return address
	}
	return fmt.Sprintf("%s..%s", address[0:4], address[len(address)-4:])
}

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

func Plural(singularForm string, value int) string {
	if value == 1 {
		return singularForm
	} else {
		return singularForm + "s"
	}
}
func PluralString(singularForm string, valueStr string) string {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Error(err)
	}
	if value == 1 {
		return singularForm
	} else {
		return singularForm + "s"
	}
}
