package errors

import (
	"context"

	"go.innotegrity.dev/mod/xerrors"
)

// BackendDeleteError is returned when a deletion of a resource fails.
type BackendDeleteError struct{ *xerrors.XError }

// NewBackendDeleteError creates a new [BackendDeleteError] object.
//
// If the error is nil, a new [BackendDeleteError] object is created with the given message and arguments.
// If the error is not nil, the [BackendDeleteError] object is created with the error wrapped and the given
// message and arguments.
//
// The context is used to set the options for the error.
func NewBackendDeleteError(ctx context.Context, err error, msg string, args ...any) *BackendDeleteError {
	return &BackendDeleteError{XError: xerrors.NewXError(ctx, err, BackendDeleteErrorCode, msg, args...)}
}

// BackendGetError is returned when retrieving the contents of a resource fails.
type BackendGetError struct{ *xerrors.XError }

// NewBackendGetError creates a new [BackendGetError] object.
//
// If the error is nil, a new [BackendGetError] object is created with the given message and arguments.
// If the error is not nil, the [BackendGetError] object is created with the error wrapped and the given
// message and arguments.
//
// The context is used to set the options for the error.
func NewBackendGetError(ctx context.Context, err error, msg string, args ...any) *BackendGetError {
	return &BackendGetError{XError: xerrors.NewXError(ctx, err, BackendGetErrorCode, msg, args...)}
}

// BackendExistsError is returned when checking for the existence of a resource fails.
type BackendExistsError struct{ *xerrors.XError }

// NewBackendExistsError creates a new [BackendExistsError] object.
//
// If the error is nil, a new [BackendExistsError] object is created with the given message and arguments.
// If the error is not nil, the [BackendExistsError] object is created with the error wrapped and the given
// message and arguments.
//
// The context is used to set the options for the error.
func NewBackendExistsError(ctx context.Context, err error, msg string, args ...any) *BackendExistsError {
	return &BackendExistsError{XError: xerrors.NewXError(ctx, err, BackendExistsErrorCode, msg, args...)}
}

// BackendInitError is returned when initializing a backend fails.
type BackendInitError struct{ *xerrors.XError }

// NewBackendInitError creates a new [BackendInitError] object.
//
// If the error is nil, a new [BackendInitError] object is created with the given message and arguments.
// If the error is not nil, the [BackendInitError] object is created with the error wrapped and the given
// message and arguments.
//
// The context is used to set the options for the error.
func NewBackendInitError(ctx context.Context, err error, msg string, args ...any) *BackendInitError {
	return &BackendInitError{XError: xerrors.NewXError(ctx, err, BackendInitErrorCode, msg, args...)}
}

// BackendListError is returned when listing a resource fails.
type BackendListError struct{ *xerrors.XError }

// NewBackendListError creates a new [BackendListError] object.
//
// If the error is nil, a new [BackendListError] object is created with the given message and arguments.
// If the error is not nil, the [BackendListError] object is created with the error wrapped and the given
// message and arguments.
//
// The context is used to set the options for the error.
func NewBackendListError(ctx context.Context, err error, msg string, args ...any) *BackendListError {
	return &BackendListError{XError: xerrors.NewXError(ctx, err, BackendListErrorCode, msg, args...)}
}

// BackendPutError is returned when writing the contents of a resource fails.
type BackendPutError struct{ *xerrors.XError }

// NewBackendPutError creates a new [BackendPutError] object.
//
// If the error is nil, a new [BackendPutError] object is created with the given message and arguments.
// If the error is not nil, the [BackendPutError] object is created with the error wrapped and the given
// message and arguments.
//
// The context is used to set the options for the error.
func NewBackendPutError(ctx context.Context, err error, msg string, args ...any) *BackendPutError {
	return &BackendPutError{XError: xerrors.NewXError(ctx, err, BackendPutErrorCode, msg, args...)}
}
