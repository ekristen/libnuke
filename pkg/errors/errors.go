// Package errors provides common errors that can be used throughout the library for handling of resource errors
package errors

import "errors"

type ErrSkipRequest string

func (err ErrSkipRequest) Error() string {
	return string(err)
}

type ErrUnknownEndpoint string

func (err ErrUnknownEndpoint) Error() string {
	return string(err)
}

type ErrWaitResource string

func (err ErrWaitResource) Error() string {
	return string(err)
}

type ErrHoldResource string

func (err ErrHoldResource) Error() string {
	return string(err)
}

// ErrFailedResource is returned by Filter() for a resource that still exists but is stuck in a state it
// cannot leave, so a wait round marks it failed instead of finished
type ErrFailedResource string

func (err ErrFailedResource) Error() string {
	return string(err)
}

type ErrUnknownPreset string

func (err ErrUnknownPreset) Error() string {
	return string(err)
}

type ErrDeprecatedResourceType string

func (err ErrDeprecatedResourceType) Error() string {
	return string(err)
}

var ErrNoBlocklistDefined = errors.New("no blocklist defined")
var ErrBlocklistAccount = errors.New("account is in blocklist")
var ErrAccountNotConfigured = errors.New("account is not configured")
