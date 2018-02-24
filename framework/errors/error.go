// Copyright 2018 XService, All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not
// use this file except in compliance with the License. A copy of the License is
// located at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// or in the "license" file accompanying this file. This file is distributed on
// an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// This file contains some code from  https://github.com/twitchtv/twirp/:
// Copyright 2018 Twitch Interactive, Inc.  All Rights Reserved.
// https://github.com/twitchtv/twirp/

package errors

import (
	"fmt"
	"net/http"
)

// Error represents an error in a service call.
type Error interface {
	// Code is of the valid error codes.
	Code() ErrorCode

	// Msg returns a human-readable, unstructured messages describing the error.
	Msg() string

	// WithMeta returns a copy of the Error with the given key-value pair attached
	// as metadata. If the key is already set, it is overwritten.
	WithMeta(key string, val string) Error

	// Meta returns the stored value for the given key. If the key has no set
	// value, Meta returns an empty string. There is no way to distinguish between
	// an unset value and an explicit empty string.
	Meta(key string) string

	// MetaMap returns the complete key-value metadata map stored on the error.
	MetaMap() map[string]string

	// Error returns a string of the form "error <Type>: <Msg>"
	Error() string
}

// NewError is the generic constructor for a Error. The ErrorCode must be
// one of the valid predefined constants, otherwise it will be converted to an
// error {type: Internal, msg: "invalid error type {{code}}"}. If you need to
// add metadata, use .WithMeta(key, value) method after building the error.
func NewError(code ErrorCode, msg string) Error {
	if IsValidErrorCode(code) {
		return &twerr{
			code: code,
			msg:  msg,
		}
	}
	return &twerr{
		code: Internal,
		msg:  "invalid error type " + string(code),
	}
}

// NotFoundError constructor for the common NotFound error.
func NotFoundError(msg string) Error {
	return NewError(NotFound, msg)
}

// InvalidArgumentError constructor for the common InvalidArgument error. Can be
// used when an argument has invalid format, is a number out of range, is a bad
// option, etc).
func InvalidArgumentError(argument string, validationMsg string) Error {
	err := NewError(InvalidArgument, argument+" "+validationMsg)
	err = err.WithMeta("argument", argument)
	return err
}

// RequiredArgumentError is a more specific constructor for InvalidArgument
// error. Should be used when the argument is required (expected to have a
// non-zero value).
func RequiredArgumentError(argument string) Error {
	return InvalidArgumentError(argument, "is required")
}

// InternalError constructor for the common Internal error. Should be used to
// specify that something bad or unexpected happened.
func InternalError(msg string) Error {
	return NewError(Internal, msg)
}

// InternalErrorWith is an easy way to wrap another error. It adds the
// underlying error's type as metadata with a key of "cause", which can be
// useful for debugging. Should be used in the common case of an unexpected
// error returned from another API, but sometimes it is better to build a more
// specific error (like with NewError(Unknown, err.Error()), for example).
//
// The returned error also has a Cause() method which will return the original
// error, if it is known. This can be used with the github.com/pkg/errors
// package to extract the root cause of an error. Information about the root
// cause of an error is lost when it is serialized, so this doesn't let a client
// know the exact root cause of a server's error.
func InternalErrorWith(err error) Error {
	msg := err.Error()
	terr := NewError(Internal, msg)
	terr = terr.WithMeta("cause", fmt.Sprintf("%T", err)) // to easily tell apart wrapped internal errors from explicit ones
	return &WrappedErr{
		wrapper: terr,
		cause:   err,
	}
}

// ErrorCode represents a  error type.
type ErrorCode string

// Valid  error types. Most error types are equivalent to gRPC status codes
// and follow the same semantics.
const (
	// Canceled indicates the operation was cancelled (typically by the caller).
	Canceled ErrorCode = "canceled"

	// Unknown error. For example when handling errors raised by APIs that do not
	// return enough error information.
	Unknown ErrorCode = "unknown"

	// InvalidArgument indicates client specified an invalid argument. It
	// indicates arguments that are problematic regardless of the state of the
	// system (i.e. a malformed file name, required argument, number out of range,
	// etc.).
	InvalidArgument ErrorCode = "invalid_argument"

	// DeadlineExceeded means operation expired before completion. For operations
	// that change the state of the system, this error may be returned even if the
	// operation has completed successfully (timeout).
	DeadlineExceeded ErrorCode = "deadline_exceeded"

	// NotFound means some requested entity was not found.
	NotFound ErrorCode = "not_found"

	// BadRoute means that the requested URL path wasn't routable to a
	// service and method. This is returned by the generated server, and usually
	// shouldn't be returned by applications. Instead, applications should use
	// NotFound or Unimplemented.
	BadRoute ErrorCode = "bad_route"

	// AlreadyExists means an attempt to create an entity failed because one
	// already exists.
	AlreadyExists ErrorCode = "already_exists"

	// PermissionDenied indicates the caller does not have permission to execute
	// the specified operation. It must not be used if the caller cannot be
	// identified (Unauthenticated).
	PermissionDenied ErrorCode = "permission_denied"

	// Unauthenticated indicates the request does not have valid authentication
	// credentials for the operation.
	Unauthenticated ErrorCode = "unauthenticated"

	// ResourceExhausted indicates some resource has been exhausted, perhaps a
	// per-user quota, or perhaps the entire file system is out of space.
	ResourceExhausted ErrorCode = "resource_exhausted"

	// FailedPrecondition indicates operation was rejected because the system is
	// not in a state required for the operation's execution. For example, doing
	// an rmdir operation on a directory that is non-empty, or on a non-directory
	// object, or when having conflicting read-modify-write on the same resource.
	FailedPrecondition ErrorCode = "failed_precondition"

	// Aborted indicates the operation was aborted, typically due to a concurrency
	// issue like sequencer check failures, transaction aborts, etc.
	Aborted ErrorCode = "aborted"

	// OutOfRange means operation was attempted past the valid range. For example,
	// seeking or reading past end of a paginated collection.
	//
	// Unlike InvalidArgument, this error indicates a problem that may be fixed if
	// the system state changes (i.e. adding more items to the collection).
	//
	// There is a fair bit of overlap between FailedPrecondition and OutOfRange.
	// We recommend using OutOfRange (the more specific error) when it applies so
	// that callers who are iterating through a space can easily look for an
	// OutOfRange error to detect when they are done.
	OutOfRange ErrorCode = "out_of_range"

	// Unimplemented indicates operation is not implemented or not
	// supported/enabled in this service.
	Unimplemented ErrorCode = "unimplemented"

	// Internal errors. When some invariants expected by the underlying system
	// have been broken. In other words, something bad happened in the library or
	// backend service. Do not confuse with HTTP Internal Server Error; an
	// Internal error could also happen on the client code, i.e. when parsing a
	// server response.
	Internal ErrorCode = "internal"

	// Unavailable indicates the service is currently unavailable. This is a most
	// likely a transient condition and may be corrected by retrying with a
	// backoff.
	Unavailable ErrorCode = "unavailable"

	// DataLoss indicates unrecoverable data loss or corruption.
	DataLoss ErrorCode = "data_loss"

	// NoError is the zero-value, is considered an empty error and should not be
	// used.
	NoError ErrorCode = ""
)

// ServerHTTPStatusFromErrorCode maps a  error type into a similar HTTP
// response status. It is used by the  server handler to set the HTTP
// response status code. Returns 0 if the ErrorCode is invalid.
func ServerHTTPStatusFromErrorCode(code ErrorCode) int {
	switch code {
	case Canceled:
		return http.StatusRequestTimeout
	case Unknown:
		return http.StatusInternalServerError
	case InvalidArgument:
		return http.StatusBadRequest
	case DeadlineExceeded:
		return http.StatusRequestTimeout
	case NotFound:
		return http.StatusNotFound
	case BadRoute:
		return http.StatusNotFound
	case AlreadyExists:
		return http.StatusConflict
	case PermissionDenied:
		return http.StatusForbidden
	case Unauthenticated:
		return http.StatusUnauthorized
	case ResourceExhausted:
		return http.StatusForbidden
	case FailedPrecondition:
		return http.StatusPreconditionFailed
	case Aborted:
		return http.StatusConflict
	case OutOfRange:
		return http.StatusBadRequest
	case Unimplemented:
		return http.StatusNotImplemented
	case Internal:
		return http.StatusInternalServerError
	case Unavailable:
		return http.StatusServiceUnavailable
	case DataLoss:
		return http.StatusInternalServerError
	case NoError:
		return http.StatusOK
	default:
		return 0 // Invalid!
	}
}

// IsValidErrorCode returns true if is one of the valid predefined constants.
func IsValidErrorCode(code ErrorCode) bool {
	return ServerHTTPStatusFromErrorCode(code) != 0
}

// .Error implementation
type twerr struct {
	code ErrorCode
	msg  string
	meta map[string]string
}

func (e *twerr) Code() ErrorCode { return e.code }
func (e *twerr) Msg() string     { return e.msg }

func (e *twerr) Meta(key string) string {
	if e.meta != nil {
		return e.meta[key] // also returns "" if key is not in meta map
	}
	return ""
}

func (e *twerr) WithMeta(key string, value string) Error {
	newErr := &twerr{
		code: e.code,
		msg:  e.msg,
		meta: make(map[string]string, len(e.meta)),
	}
	for k, v := range e.meta {
		newErr.meta[k] = v
	}
	newErr.meta[key] = value
	return newErr
}

func (e *twerr) MetaMap() map[string]string {
	return e.meta
}

func (e *twerr) Error() string {
	return fmt.Sprintf(" error %s: %s", e.code, e.msg)
}

// wrappedErr fulfills the .Error interface and the
// github.com/pkg/errors.Causer interface. It exposes all the  error
// methods, but root cause of an error can be retrieved with
// (*wrappedErr).Cause. This is expected to be used with the InternalErrorWith
// function.
type WrappedErr struct {
	wrapper Error
	cause   error
}

func (e *WrappedErr) Code() ErrorCode            { return e.wrapper.Code() }
func (e *WrappedErr) Msg() string                { return e.wrapper.Msg() }
func (e *WrappedErr) Meta(key string) string     { return e.wrapper.Meta(key) }
func (e *WrappedErr) MetaMap() map[string]string { return e.wrapper.MetaMap() }
func (e *WrappedErr) Error() string              { return e.wrapper.Error() }
func (e *WrappedErr) WithMeta(key string, val string) Error {
	return &WrappedErr{
		wrapper: e.wrapper.WithMeta(key, val),
		cause:   e.cause,
	}
}
func (e *WrappedErr) Cause() error { return e.cause }

// wrappedError implements the github.com/pkg/errors.Causer interface, allowing errors to be
// examined for their root cause.
type wrappedError struct {
	msg   string
	cause error
}

func WrapErr(err error, msg string) error { return &wrappedError{msg: msg, cause: err} }
func (e *wrappedError) Cause() error      { return e.cause }
func (e *wrappedError) Error() string     { return e.msg + ": " + e.cause.Error() }

// ClientError adds consistency to errors generated in the client
func ClientError(desc string, err error) Error {
	return InternalErrorWith(WrapErr(err, desc))
}

// badRouteError is used when the twirp server cannot route a request
func BadRouteError(msg string, method, url string) Error {
	err := NewError(BadRoute, msg)
	err = err.WithMeta("xservice_invalid_route", method+" "+url)
	return err
}
