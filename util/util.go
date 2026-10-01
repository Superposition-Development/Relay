package util

func RemoveByValue[T comparable](s []T, value T) []T {
	for i, v := range s {
		if v == value {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func TruncateStrings(strings []string, maxLength int) []string {
	for i, s := range strings {
		runes := []rune(s)

		if len(runes) > maxLength {
			strings[i] = string(runes[:maxLength])
		}
	}

	return strings
}

func Clamp(val, minVal, maxVal int) int {
	if val < minVal {
		return minVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

//this is probably a pointless package...
