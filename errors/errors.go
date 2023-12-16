package errors

import (
	"fmt"
	"net/http"
)

const (
	UnknownCode    = -1
	UnknownReason  = "UnknownError"
	UnknownMessage = "unknown error"
	DefaultStatus  = http.StatusInternalServerError
)

type Error interface {
	error
	Code() int32
	HttpStatus() int32
	Reason() string
	Message() string
	Metadata() map[string]string
	Unwrap() error
}

type ErrorImpl struct {
	cause     error
	status    int32
	Code_     int32             `json:"code"`
	Reason_   string            `json:"reason"`
	Message_  string            `json:"message"`
	Metadata_ map[string]string `json:"metadata"`
}

func (e *ErrorImpl) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v cause = %v",
			e.Code_, e.Reason_, e.Message_, e.Metadata(), e.cause)
	}

	return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v",
		e.Code_, e.Reason_, e.Message_, e.Metadata())
}

func (e *ErrorImpl) Code() int32 {
	return e.Code_
}

func (e *ErrorImpl) HttpStatus() int32 {
	return e.status
}

func (e *ErrorImpl) Reason() string {
	return e.Reason_
}

func (e *ErrorImpl) Message() string {
	return e.Message_
}

func (e *ErrorImpl) Metadata() map[string]string {
	return e.Metadata_
}

func (e *ErrorImpl) Unwrap() error {
	return e.cause
}

func New(code, status int32, reason, message string) Error {
	return &ErrorImpl{
		status:   status,
		Code_:    code,
		Reason_:  reason,
		Message_: message,
	}
}

func Newf(code, status int32, reason, format string, args ...interface{}) Error {
	return Newf(code, status, reason, fmt.Sprintf(format, args...))
}

func FromError(code, status int32, reason, message string, err error) Error {
	return &ErrorImpl{
		cause:    err,
		status:   status,
		Code_:    code,
		Reason_:  reason,
		Message_: message,
	}
}

func FromErrorf(code, status int32, reason string, err error, format string, args ...interface{}) Error {
	return FromError(code, status, reason, fmt.Sprintf(format, args...), err)
}

func IsError(err error) bool {
	_, ok := err.(Error)

	return ok
}
