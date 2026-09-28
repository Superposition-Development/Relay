package util

func RemoveByValue[T comparable](s []T, value T) []T {
	for i, v := range s {
		if v == value {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

//this is probably a pointless package...
