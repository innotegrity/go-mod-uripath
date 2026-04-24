package errors

import (
	"context"

	"go.innotegrity.dev/mod/xerrors"
)

// InvalidSchemeError is returned when the format of a scheme's name is invalid.
type InvalidSchemeError struct{ *xerrors.XError }

// NewInvalidSchemeError creates a new [InvalidSchemeError] object.
//
// The context is used to set the options for the error.
func NewInvalidSchemeError(ctx context.Context, scheme string) *InvalidSchemeError {
	xerr := xerrors.NewXError(ctx, nil, InvalidSchemeErrorCode,
		"%s: scheme is invalid", scheme)
	_ = xerr.WithAttrs(map[string]any{
		"scheme": scheme,
	})

	return &InvalidSchemeError{XError: xerr}
}

// SchemeNotFoundError is returned when a scheme is not registered.
type SchemeNotFoundError struct{ *xerrors.XError }

// NewSchemeNotFoundError creates a new [SchemeNotFoundError] object.
//
// The context is used to set the options for the error.
func NewSchemeNotFoundError(ctx context.Context, scheme string) *SchemeNotFoundError {
	xerr := xerrors.NewXError(ctx, nil, SchemeNotFoundErrorCode,
		"%s: scheme not found", scheme)
	_ = xerr.WithAttrs(map[string]any{
		"scheme": scheme,
	})

	return &SchemeNotFoundError{XError: xerr}
}

// SchemeExistsError is returned when a scheme is already registered.
type SchemeExistsError struct{ *xerrors.XError }

// NewSchemeExistsError creates a new [SchemeExistsError] object.
//
// The context is used to set the options for the error.
func NewSchemeExistsError(ctx context.Context, scheme string) *SchemeExistsError {
	xerr := xerrors.NewXError(ctx, nil, SchemeExistsErrorCode,
		"%s: scheme is already registered", scheme)
	_ = xerr.WithAttrs(map[string]any{
		"scheme": scheme,
	})

	return &SchemeExistsError{XError: xerr}
}
