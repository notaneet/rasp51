package utils

import (
	"regexp"
	"strconv"
)

var spaceRE = regexp.MustCompile(`\s+`)

// RemoveSpaces удалить повторяющиеся пробелы
func RemoveSpaces(s string) string {
	return spaceRE.ReplaceAllString(s, " ")
}

// MustAtoi Конвертация стриги в число (строгое)
func MustAtoi(str string) int {
	ret, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return ret
}
