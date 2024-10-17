package accounts

// hidden returns 0 when the width is greater than the fill
func hidden(width int, fillSize int) int {
	if fillSize < width {
		return 0
	}
	return width
}
