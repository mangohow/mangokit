package parse

import (
	"fmt"
	"strconv"
)

func MustParseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("parse int error, string: %s", s))
	}
	return n
}
