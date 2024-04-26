package x1

// flatBytes creates a bit slice from a byte slice by converting each bit in every byte to a boolean.
func flatBytes(bytes []byte) []bool {
	bools := []bool{}
	for i := range bytes {
		for j := 0; j < 8; j++ {
			result := ((bytes[i] >> j) & 1) != 0 // true / false based on the button state.
			bools = append(bools, result)
		}
	}

	return bools
}
