// Package uripath maps resource locations to [Backend] implementations using URI-shaped strings
// (scheme, host, path, query, and fragment). Callers parse a URI with [ParseURI], then invoke
// methods on [URI] such as [URI.Get], [URI.Put], [URI.Delete], [URI.Exists], and [URI.List]—each
// delegates to the backend selected by the scheme.
//
// # Registration
//
// Schemes are resolved through a process-wide registry. Use [RegisterBackend] to associate a
// scheme name with a [NewBackendFunc] constructor. Built-in implementations live under
// [go.innotegrity.dev/mod/uripath/backends]; register them with a blank import so their
// [RegisterBackend] calls run at startup, for example:
//
//	import (
//		"context"
//
//		"go.innotegrity.dev/mod/uripath"
//
//		_ "go.innotegrity.dev/mod/uripath/backends/aws"
//		_ "go.innotegrity.dev/mod/uripath/backends/generic"
//	)
//
// Custom backends typically embed [BackendBase] to satisfy [Backend] with defaults and to reuse
// common option handling; override the methods you need.
//
// # Context and options
//
// [ParseURI], [RegisterBackend], and [BackendAs] take a [context.Context] first. [ParseURI] also
// accepts optional [BackendOption] values to pass configuration when the URI string alone is not
// enough. Backends may read options from query parameters and/or from these options; see each
// backend’s documentation.
//
// # Errors
//
// API functions return [go.innotegrity.dev/mod/xerrors.Error] objects.  See the
// [go.innotegrity.dev/mod/uripath/errors] package for more information.  If you wish to pass options for errors
// to the API functions, you can the context to do so.  Otherwise, you can simply pass [context.Background()] as
// the context.
//
// # Serialization
//
// [URI] supports [encoding/json] and [encoding.TextMarshaler] via [URI.MarshalJSON], [URI.UnmarshalJSON],
// [URI.MarshalText], and [URI.UnmarshalText]. [URI.SafeString] applies the same redaction without JSON
// or text encoding. Unmarshaling marshaled output does not restore the original username or password.
// Use [URI.String] when the full URI, including secrets, is required in memory.
package uripath
