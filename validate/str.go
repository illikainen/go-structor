package validate

func isPrintable[T []byte | string](s T) bool {
	var input []byte
	if value, ok := any(s).([]byte); ok {
		input = value
	} else {
		input = []byte(s)
	}

	for _, b := range input {
		if b != 0x0a && (b < 0x20 || b > 0x7e) {
			return false
		}
	}

	return true
}
