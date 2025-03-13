package log

import (
	"fmt"
	"os"
	"strings"
)

var debug = true

func Debugf(format string, args ...interface{}) {
	if !debug {
		return
	}
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Printf(format, args...)
}

func Logf(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Printf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

func Fatalf(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
