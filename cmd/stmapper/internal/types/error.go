package types

import "fmt"

type UnsupportedTypeConversionError struct {
	File  string
	Fn    string
	Param string
}

func NewUnsupportedTypeConversionError(file string, fn string, param string) *UnsupportedTypeConversionError {
	return &UnsupportedTypeConversionError{File: file, Param: param, Fn: fn}
}

func (u *UnsupportedTypeConversionError) Error() string {
	return fmt.Sprintf("unsupported type conversion, in file %s, func %s, parameter %s", u.File, u.Fn, u.Param)
}

type MultiOutputParameterError struct {
	File string
	Fn   string
}

func NewMultiOutputParameterError(fn string, file string) *MultiOutputParameterError {
	return &MultiOutputParameterError{Fn: fn, File: file}
}

func (m *MultiOutputParameterError) Error() string {
	return fmt.Sprintf("unsupported multi output parameter, in file %s, func %s", m.File, m.Fn)
}
