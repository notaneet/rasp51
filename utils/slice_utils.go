package utils

// GetOrString получить i-тый элемент в slice, если нет, то вернуть or
func GetOrString(slice []string, i int, or string) string {
	if len(slice)-1 >= i {
		return slice[i]
	}
	return or
}
