package service

import "errors"

var (
	ErrValidation    = errors.New("validation error")
	ErrConfiguration = errors.New("configuration error")
	ErrNotFound      = errors.New("not found")
)
