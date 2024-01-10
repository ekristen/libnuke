// Package errors provides common errors that can be used throughout the library for handling of resource errors
package errors

type ErrSkipRequest string

func (err ErrSkipRequest) Error() string {
	return string(err)
}

type ErrUnknownEndpoint string

func (err ErrUnknownEndpoint) Error() string {
	return string(err)
}
