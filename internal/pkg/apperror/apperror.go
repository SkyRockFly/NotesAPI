package apperror

import "errors"

var (
	ErrInvalidJSON   = errors.New("invalid json")   // 422
	ErrBadRequest    = errors.New("bad request")    // 400
	ErrNotFound      = errors.New("not found")      // 404
	ErrBackend       = errors.New("gateway error")  // 502
	ErrAlreadyExists = errors.New("already exists") // 409
	ErrUnauthorized  = errors.New("unauthorized")   // 401
)
