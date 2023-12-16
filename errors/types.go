package errors

import "net/http"

func BadRequest(code int32, reason, message string) Error {
	return New(code, http.StatusBadRequest, reason, message)
}

func BadRequestCause(code int32, reason, message string, err error) Error {
	return FromError(code, http.StatusBadRequest, reason, message, err)
}

func Unauthorized(code int32, reason, message string) Error {
	return New(code, http.StatusUnauthorized, reason, message)
}

func UnauthorizedCause(code int32, reason, message string, err error) Error {
	return FromError(code, http.StatusUnauthorized, reason, message, err)
}

func Forbidden(code int32, reason, message string) Error {
	return New(code, http.StatusForbidden, reason, message)
}

func ForbiddenCause(code int32, reason, message string, err error) Error {
	return FromError(code, http.StatusForbidden, reason, message, err)
}

func NotFound(code int32, reason, message string) Error {
	return New(code, http.StatusNotFound, reason, message)
}

func NotFoundCause(code int32, reason, message string, err error) Error {
	return FromError(code, http.StatusNotFound, reason, message, err)
}

func InternalServer(code int32, reason, message string) Error {
	return New(code, http.StatusInternalServerError, reason, message)
}

func InternalServerCause(code int32, reason, message string, err error) Error {
	return FromError(code, http.StatusInternalServerError, reason, message, err)
}

func ServiceUnavailable(code int32, reason, message string) Error {
	return New(code, http.StatusServiceUnavailable, reason, message)
}

func ServiceUnavailableCause(code int32, reason, message string, err error) Error {
	return FromError(code, http.StatusServiceUnavailable, reason, message, err)
}
