package errors

import (
	"context"

	"go.innotegrity.dev/mod/xerrors"
)

// InvalidParameterError is returned when an invalid parameter is passed to a function.
type InvalidParameterError struct{ *xerrors.XError }

// NewInvalidParameterError creates a new [InvalidParameterError] object.
//
// The context is used to set the options for the error.
func NewInvalidParameterError(ctx context.Context, err error, message string, args ...any) *InvalidParameterError {
	return &InvalidParameterError{XError: xerrors.NewXError(ctx, err, InvalidParameterErrorCode, message, args...)}
}
