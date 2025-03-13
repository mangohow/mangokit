package strutil

import (
	"math/rand/v2"
	"strings"
)

const (
	lowerLetter = "abcdefghijklmnopqrstuvwxyz"
	upperLetter = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letter      = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func RandStringCustom(length int, letter string) string {
	builder := strings.Builder{}
	builder.Grow(length)
	for i := 0; i < length; i++ {
		builder.WriteByte(letter[rand.IntN(len(letter))])
	}
	return builder.String()
}

func RandLowerString(length int) string {
	return RandStringCustom(length, lowerLetter)
}

func RandUpperString(length int) string {
	return RandStringCustom(length, upperLetter)
}

func RandString(length int) string {
	return RandStringCustom(length, letter)
}

func CutBefore(s, sep string) (string, bool) {
	res, _, found := strings.Cut(s, sep)
	return res, found
}

func CutAfter(s, sep string) (string, bool) {
	_, res, found := strings.Cut(s, sep)
	return res, found
}
