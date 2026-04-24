package uripath

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"sync"

	urierrors "go.innotegrity.dev/mod/uripath/errors"
	"go.innotegrity.dev/mod/xerrors"
)

var (
	//nolint:gochecknoglobals // this is a global registry for backend schemes.
	_backendRegistry = make(map[string]NewBackendFunc)

	//nolint:gochecknoglobals // this is a global mutex for the backend registry.
	_registryMutex sync.RWMutex
)

// BackendAs returns the backend cast to the concrete type T.
//
// This function may return an error with any of the following codes:
//   - [urierrors.InvalidParameterError]: the backend does not match the given type
//
//nolint:ireturn // want to be able to return generic type.
func BackendAs[T Backend](ctx context.Context, u *URI) (T, xerrors.Error) {
	backend, ok := u.Backend.(T)
	if !ok {
		var zero T

		return zero, urierrors.NewInvalidParameterError(ctx, nil, "backend is not of type %T", zero).
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
// This function may return one of the following errors:
//   - [urierrors.InvalidSchemeError]: the scheme is not valid
//   - [urierrors.SchemeExistsError]: the scheme is already registered (and overwriteExisting was false or
//     not specified)
func RegisterBackend(ctx context.Context, scheme string, newBackendFunc NewBackendFunc, overwriteExisting ...bool,
) xerrors.Error {
	_registryMutex.Lock()
	defer _registryMutex.Unlock()

	// remove any trailing '://' or ':'characters passed by mistake
	scheme = strings.ToLower(scheme)
	scheme = strings.TrimSuffix(scheme, "://")
	scheme = strings.TrimSuffix(scheme, ":")

	// ensure the scheme is valid
	xerr := validateScheme(ctx, scheme)
	if xerr != nil {
		return xerr
	}

	allowOverwrite := len(overwriteExisting) > 0 && overwriteExisting[0]
	if _, exists := _backendRegistry[scheme]; exists && !allowOverwrite {
		return urierrors.NewSchemeExistsError(ctx, scheme)
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
// This function may return one of the following errors:
//   - [urierrors.InvalidParameterError]: the URI is not valid
//   - [urierrors.SchemeNotFoundError]: the URI scheme is not registered
func ParseURI(ctx context.Context, uri string, backendOptions ...BackendOption) (*URI, xerrors.Error) {
	// parse the URL
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, urierrors.NewInvalidParameterError(ctx, err, "failed to parse URI '%s': %s", uri, err.Error()).
			WithAttr("uri", uri)
	}

	uriObj := &URI{
		scheme:   strings.ToLower(parsedURL.Scheme),
		Host:     parsedURL.Host,
		Path:     parsedURL.Path,
		Query:    parsedURL.Query(),
		Fragment: parsedURL.Fragment,
	}
	if parsedURL.User != nil {
		uriObj.Username = parsedURL.User.Username()
		uriObj.Password, _ = parsedURL.User.Password()
	}

	// construct the backend
	newBackendFunc, xerr := getBackend(ctx, uriObj.scheme)
	if xerr != nil {
		return nil, xerr
	}

	uriObj.Backend, xerr = newBackendFunc(ctx, uriObj, backendOptions...)
	if xerr != nil {
		return nil, xerr
	}

	return uriObj, nil
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
// Users can also register custom backends for additional schemes using the [RegisterBackend] function.
//
// User info fields [URI.Username] and [URI.Password] hold the values from parsing or construction.
// [URI.String] includes them verbatim; [URI.SafeString], [URI.MarshalJSON], and [URI.MarshalText] emit
// redacted values that cannot be reversed by [URI.UnmarshalJSON] or [URI.UnmarshalText].
//
//nolint:recvcheck // need to abide by json.Marshaler and json.Unmarshaler interfaces.
type URI struct {
	// Backend holds the backend which actually handles doing the work.
	Backend Backend

	// Fragment holds the portion of the URI after the pound (#) character (if there is any).
	Fragment string

	// Host holds the hostname (and optional port) of the URI.
	Host string

	// Password holds the password portion of the URI (if one was specified).
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

// MarshalJSON marshals [URI.SafeString] as a JSON string.
//
// Unmarshaling that JSON with [URI.UnmarshalJSON] yields a URI whose user info is the redacted literal string,
// not the original secret values.
func (u URI) MarshalJSON() ([]byte, error) {
	result, err := json.Marshal(u.redactedString())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal URI to JSON: %w", err)
	}

	return result, nil
}

// MarshalText marshals [URI.SafeString] to plain text.
//
// Unmarshaling that text with [URI.UnmarshalText] yields a URI whose user info is the redacted literal string,
// not the original secret values.
func (u URI) MarshalText() ([]byte, error) {
	return []byte(u.redactedString()), nil
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

// SafeString returns a URI string with credentials redacted.
//
// The username keeps only the first and last Unicode code point (the rest are '*'), and the password is
// replaced by a set of '*' characters. It returns the empty string if u is nil.
func (u *URI) SafeString() string {
	if u == nil {
		return ""
	}

	return u.redactedString()
}

// Scheme returns the URI scheme (e.g., "s3", "file", "git").
func (u *URI) Scheme() string {
	return u.scheme
}

// String returns the full URI string including the original username and password.
//
// For a redacted form, use [URI.SafeString], [URI.MarshalJSON], or [URI.MarshalText].
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
//
// If the JSON was produced by [URI.MarshalJSON], the embedded credentials are the redacted placeholders only;
// the original username and password cannot be recovered.
func (u *URI) UnmarshalJSON(data []byte) error {
	var sval string

	err := json.Unmarshal(data, &sval)
	if err != nil {
		return fmt.Errorf("failed to unmarshal URI from JSON: %w", err)
	}

	val, err := ParseURI(context.Background(), sval)
	if err != nil {
		return err
	}

	*u = *val

	return nil
}

// UnmarshalText parses the text into a [URI] value.
//
// Text produced by [URI.MarshalText] contains redacted credentials and will not round-trip to the original
// secret values.
func (u *URI) UnmarshalText(data []byte) error {
	val, err := ParseURI(context.Background(), string(data))
	if err != nil {
		return err
	}

	*u = *val

	return nil
}

// redactedString returns a URI string like [URI.String] but with credentials redacted for safe
// logging or serialization. It backs [URI.SafeString], [URI.MarshalJSON], and [URI.MarshalText].
func (u URI) redactedString() string {
	var builder strings.Builder
	builder.WriteString(u.scheme)
	builder.WriteString("://")

	if u.Username != "" {
		builder.WriteString(redactUsernameForMarshal(u.Username))

		if u.Password != "" {
			builder.WriteString(":")
			builder.WriteString(redactedPasswordPlaceholder)
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

// redactedPasswordPlaceholder is a fixed run of asterisks used in place of the real password when
// serializing a URI with [URI.MarshalJSON] or [URI.MarshalText], so the password length is not
// disclosed.
const redactedPasswordPlaceholder = "********"

// redactUsernameForMarshal returns a redacted form of username suitable for marshaling: the first
// and last Unicode code points are kept; any others are replaced with '*'. A single code point is
// replaced entirely by '*'; two code points are left unchanged.
func redactUsernameForMarshal(username string) string {
	char := []rune(username)
	switch length := len(char); length {
	case 0:
		return ""
	case 1:
		return "*"
	//nolint:mnd // need to return the first and last characters of the username.
	case 2:
		return string(char[0]) + string(char[1])
	default:
		//nolint:mnd // need to repeat the asterisk character length-2 times.
		return string(char[0]) + strings.Repeat("*", length-2) + string(char[length-1])
	}
}

// getBackend returns the backend factory function for a given scheme.
//
// This function may return an error with any of the following codes:
//   - [urierrors.SchemeNotFoundError]: the scheme is not registered
func getBackend(ctx context.Context, scheme string) (NewBackendFunc, xerrors.Error) {
	_registryMutex.RLock()
	defer _registryMutex.RUnlock()

	if newBackendFunc, exists := _backendRegistry[strings.ToLower(scheme)]; exists {
		return newBackendFunc, nil
	}

	return nil, urierrors.NewSchemeNotFoundError(ctx, scheme)
}

// validateScheme checks to ensure the given scheme is valid.
//
// This may return one of the following errors:
//   - [urierrors.InvalidSchemeError]: the scheme is not valid
func validateScheme(ctx context.Context, scheme string) xerrors.Error {
	if !regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]+$`).MatchString(scheme) {
		return urierrors.NewInvalidSchemeError(ctx, scheme)
	}

	return nil
}
