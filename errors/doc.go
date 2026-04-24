// Package errors defines typed errors and numeric codes used by [go.innotegrity.dev/mod/uripath]
// and its backends.
//
// # Representation
//
// Each exported error type embeds [*go.innotegrity.dev/mod/xerrors.XError] and implements
// [go.innotegrity.dev/mod/xerrors.Error]. Use [go.innotegrity.dev/mod/xerrors.Error.Code] for
// comparisons against the *ErrorCode constants in this package, or use [errors.As] with the
// concrete pointer types when you need typed handling.
//
// Because this package is named errors, importers usually use an alias to avoid clashing with
// the standard library, for example:
//
//	import urierrors "go.innotegrity.dev/mod/uripath/errors"
//
// # Where these errors originate
//
// Core APIs in package [go.innotegrity.dev/mod/uripath] document which constructors apply.
// Typical cases include:
//   - [go.innotegrity.dev/mod/uripath.ParseURI] — invalid URIs, unknown schemes
//   - [go.innotegrity.dev/mod/uripath.RegisterBackend] — invalid scheme names, duplicate registration
//   - [go.innotegrity.dev/mod/uripath.BackendAs] — backend type mismatch
//
// Backend implementations may also return errors whose code matches the backend-related constants
// below (for example when wrapping low-level failures during Get, Put, or initialization).
//
// # Error kinds
//
// General validation:
//   - [InvalidParameterError] — [InvalidParameterErrorCode]
//
// Scheme naming and registration:
//   - [InvalidSchemeError] — [InvalidSchemeErrorCode]
//   - [SchemeNotFoundError] — [SchemeNotFoundErrorCode]
//   - [SchemeExistsError] — [SchemeExistsErrorCode]
//
// Backend operations (constructors in this package use the matching *ErrorCode value):
//   - [BackendDeleteError] — [BackendDeleteErrorCode]
//   - [BackendGetError] — [BackendGetErrorCode]
//   - [BackendListError] — [BackendListErrorCode]
//   - [BackendPutError] — [BackendPutErrorCode]
//   - [BackendInitError] — [BackendInitErrorCode]
//
// [BackendExistsErrorCode] is reserved for failures while checking resource existence; backends
// may report it using [go.innotegrity.dev/mod/xerrors] helpers with that code.
package errors
