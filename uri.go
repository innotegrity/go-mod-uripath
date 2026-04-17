package uripath

import (
	"context"
	"encoding/json"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"go.innotegrity.dev/mod/xerrors"
)

var (
	// global registry for backend schemes
	_backendRegistry = make(map[string]NewBackendFunc)
	_registryMutex   sync.RWMutex
)

// BackendAs returns the backend cast to the concrete type T.
//
// This function may return an error with any of the following codes:
//   - [InvalidParameter]: the backend does not match the given type
func BackendAs[T Backend](u *URI) (T, xerrors.Error) {
	backend, ok := u.Backend.(T)
	if !ok {
		var zero T
		return zero, xerrors.Newf(InvalidParameter, "backend is not of type %T", zero).
			WithAttr("type", reflect.TypeOf(zero))
	}
	return backend, nil
}

// RegisterBackend registers a backend for a specific scheme, allowing users to extend the functionality with
// custom backends.
//
// Schemes are **not** case-sensitive and must abide by the following rules:
//   - They must be at least 2 characters long.
//   - They must start with an alphabetic character.
//   - They may contain letters, digits, the plus (+) sign, a period (.) or hyphen (-).
//
// If a backend is already registered for the given scheme, an error is returned unless overwriteExisting is
// explicitly set to true. Backends must be registered before calling [ParseURI] with a URI that uses the
// corresponding scheme.
//
// This function may return an error with any of the following codes:
//   - [InvalidParameter]: the scheme is not valid
//   - [SchemeExists]: the scheme is already registered (and overwriteExisting was false or not specified)
func RegisterBackend(scheme string, newBackendFunc NewBackendFunc, overwriteExisting ...bool) xerrors.Error {
	_registryMutex.Lock()
	defer _registryMutex.Unlock()

	// remove any trailing '://' or ':'characters passed by mistake
	scheme = strings.ToLower(scheme)
	scheme = strings.TrimSuffix(scheme, "://")
	scheme = strings.TrimSuffix(scheme, ":")

	// ensure the scheme is valid
	if xerr := validateScheme(scheme); xerr != nil {
		return xerr
	}

	allowOverwrite := len(overwriteExisting) > 0 && overwriteExisting[0]
	if _, exists := _backendRegistry[scheme]; exists && !allowOverwrite {
		return xerrors.Newf(SchemeExists, "%s: scheme is already registered", scheme).WithAttr("scheme", scheme)
	}
	_backendRegistry[scheme] = newBackendFunc
	return nil
}

// ParseURI parses a URI string and creates a [URI] with the appropriate backend.
//
// The uri specified **must** match a supported / registered scheme and be properly formatted.  In other words, don't
// just pass a file path directly to this function.  You'll need to use the file:// scheme for specifying file paths.
//
// Typically options should be passed to the backend as query parameters in the URI.  However, there may be cases
// where this is not desirable so the backendOptions parameter can be used to pass options directly to the backend.
// Review the individual backend documentation for how options are used during construction.
//
// The [backends] package contains the collection of built-in backends that are supported. Refer to the documentation
// for each backend for proper URI formatting.
//
// To register built-in backends, you **must** import one or more of the sub-packages in the [backends] package:
//
//		 import (
//	  	 _ "go.innotegrity.dev/mod/uripath/backends/aws"
//	  	 _ "go.innotegrity.dev/mod/uripath/backends/generic"
//	  	 _ "go.innotegrity.dev/mod/uripath/backends/hashicorp"
//	  )
//
// This function may return an error with any of the following codes:
//   - [InvalidParameter]: the URI is not valid
//   - [SchemeNotFound]: the URI scheme is not registered
func ParseURI(uri string, backendOptions ...BackendOption) (*URI, xerrors.Error) {
	// parse the URL
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, xerrors.Wrapf(InvalidParameter, err, "failed to parse URI '%s': %s", uri, err.Error()).
			WithAttr("uri", uri)
	}
	u := &URI{
		scheme:   strings.ToLower(parsedURL.Scheme),
		Host:     parsedURL.Host,
		Path:     parsedURL.Path,
		Query:    parsedURL.Query(),
		Fragment: parsedURL.Fragment,
	}
	if parsedURL.User != nil {
		u.Username = parsedURL.User.Username()
		u.Password, _ = parsedURL.User.Password()
	}

	// construct the backend
	newBackendFunc, xerr := getBackend(u.scheme)
	if xerr != nil {
		return nil, xerr
	}
	u.Backend, xerr = newBackendFunc(u, backendOptions...)
	if xerr != nil {
		return nil, xerr
	}
	return u, nil
}

// URI is used for retrieving and potentially manipulating files using a URI-style string.
//
// The URI should be in the format:
// scheme://[username[:password]@]host/path[?query][#fragment]
//
// The scheme determines which backend will be used to perform operations on the resource specified by the URI.
//
// The 'backends' package contains the collection of built-in backends that are supported. Refer to the documentation
// for each backend for proper URI formatting.
//
// To register the built-in backends, you **must** import the 'backends' package:
//
//		 import (
//	  	 _ "go.innotegrity.dev/mod/uripath/backends"
//	  )
//
// Users can also register custom backends for additional schemes using the [RegisterBackend] function.
type URI struct {
	// Backend holds the backend which actually handles doing the work.
	Backend Backend

	// Fragment holds the portion of the URI after the pound (#) character (if there is any).
	Fragment string

	// Host holds the hostname (and optional port) of the URI.
	Host string

	// Pasword holds the password portion of the URI (if one was specified).
	Password string

	// Path holds the path portion of the URI.
	Path string

	// Query holds the query parameters passed in the URI (if any).
	Query url.Values

	// Username holds the username portion of the URI (if one was specified).
	Username string

	// unexported variables
	scheme string // scheme portion of URI (before '://')
}

// Delete removes the resource at the given path.
//
// The context and options passed to this function are passed directly to the backend. Be sure
// to review the documentation for the backend to understand how the context and options are used.
//
// This function may return an error with any of the codes from the [Backend.Delete] method.
func (u *URI) Delete(ctx context.Context, options ...BackendOption) xerrors.Error {
	return u.Backend.Delete(ctx, options...)
}

// Exists checks if the resource exists at the given path.
//
// The context and options passed to this function are passed directly to the backend. Be sure
// to review the documentation for the backend to understand how the context and options are used.
//
// This function may return an error with any of the codes from the [Backend.Exists] method.
func (u *URI) Exists(ctx context.Context, options ...BackendOption) (bool, xerrors.Error) {
	return u.Backend.Exists(ctx, options...)
}

// Get retrieves the content of the resource at the given path.
//
// The context and options passed to this function are passed directly to the backend. Be sure
// to review the documentation for the backend to understand how the context and options are used.
//
// This function may return an error with any of the codes from the [Backend.Get] method.
func (u *URI) Get(ctx context.Context, options ...BackendOption) ([]byte, xerrors.Error) {
	return u.Backend.Get(ctx, options...)
}

// List lists the resources at the given path.
//
// The context and options passed to this function are passed directly to the backend. Be sure
// to review the documentation for the backend to understand how the context and options are used.
//
// This function may return an error with any of the codes from the [Backend.List] method.
func (u *URI) List(ctx context.Context, recurse bool, options ...BackendOption) ([]string, xerrors.Error) {
	return u.Backend.List(ctx, recurse, options...)
}

// MarshalJSON marshals the [URI] value to JSON.
func (u URI) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

// MarshalText marshals the [URI] value to plain text.
func (u URI) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// Put stores content at the given path.
//
// The context and options passed to this function are passed directly to the backend. Be sure
// to review the documentation for the backend to understand how the context and options are used.
//
// This function may return an error with any of the codes from the [Backend.Put] method.
func (u *URI) Put(ctx context.Context, data []byte, options ...BackendOption) xerrors.Error {
	return u.Backend.Put(ctx, data, options...)
}

// RawQuery returns the raw query string with the query parameters sorted by key.
func (u *URI) RawQuery() string {
	return u.Query.Encode()
}

// Scheme returns the URI scheme (e.g., "s3", "file", "git").
func (u *URI) Scheme() string {
	return u.scheme
}

// String returns the full URI string.
func (u *URI) String() string {
	var builder strings.Builder
	builder.WriteString(u.scheme)
	builder.WriteString("://")

	if u.Username != "" {
		builder.WriteString(u.Username)
		if u.Password != "" {
			builder.WriteString(":")
			builder.WriteString(u.Password)
		}
		builder.WriteString("@")
	}

	builder.WriteString(u.Host)
	builder.WriteString(u.Path)

	if len(u.Query) > 0 {
		builder.WriteString("?")
		builder.WriteString(u.RawQuery())
	}

	if u.Fragment != "" {
		builder.WriteString("#")
		builder.WriteString(u.Fragment)
	}

	return builder.String()
}

// UnmarshalJSON parses the JSON data into a [URI] value.
func (u *URI) UnmarshalJSON(data []byte) error {
	var sval string
	if err := json.Unmarshal(data, &sval); err != nil {
		return err
	}
	val, err := ParseURI(sval)
	if err != nil {
		return err
	}
	*u = *val
	return nil
}

// UnmarshalText parses the text into a [URI] value.
func (u *URI) UnmarshalText(data []byte) error {
	val, err := ParseURI(string(data))
	if err != nil {
		return err
	}
	*u = *val
	return nil
}

// getBackend returns the backend factory function for a given scheme.
//
// This function may return an error with any of the following codes:
//   - [SchemeNotFound]: the scheme is not registered
func getBackend(scheme string) (NewBackendFunc, xerrors.Error) {
	_registryMutex.RLock()
	defer _registryMutex.RUnlock()

	if newBackendFunc, exists := _backendRegistry[strings.ToLower(scheme)]; exists {
		return newBackendFunc, nil
	}
	return nil, xerrors.Newf(SchemeNotFound, "%s: scheme not found", scheme).WithAttr("scheme", scheme)
}

// validateScheme checks to ensure the given scheme is valid.
//
// This function may return an error with any of the following codes:
//   - [InvalidParameter]: the scheme is not valid
func validateScheme(scheme string) xerrors.Error {
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]+$`).MatchString(scheme) {
		return xerrors.Newf(InvalidParameter, "%s: scheme is invalid", scheme).WithAttr("scheme", scheme)
	}
	return nil
}
