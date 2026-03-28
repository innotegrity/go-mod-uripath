package uripath

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"go.innotegrity.dev/mod/xerrors"
)

var (
	// global registry for backend schemes
	_backendRegistry = make(map[string]NewURIPathBackendFunc)
	_registryMutex   sync.RWMutex
)

// URIPath is used for retrieving and potentially manipulating files using a URI-style string.
//
// The URI should be in the format:
// scheme://[username[:password]@]host/path[?query][#fragment]
//
// The scheme determines which backend will be used to perform operations on the resource specified by the URI.
//
// The following schemes are supported by default:
//   - abfs: for Azure Blob Storage (e.g. abfs://container/blob)
//   - awssm: for AWS Secrets Manager (e.g. awssm://secret/path?version=1)
//   - awsssm: for AWS Systems Manager (e.g. awsssm://parameter/name?version=1)
//   - azkv: for Azure Key Vault (e.g. azkv://vault/secret?version=1)
//   - file: for local filesystem operations (e.g. file:///path/to/file)
//   - hcpvault: for HashiCorp Vault secrets (e.g. hcpvault://secret/path?version=1)
//   - gcs: for Google Cloud Storage (e.g. gcs://bucket/object)
//   - git: for Git repositories (e.g. git://github.com/user/repo.git)
//   - gsm: for Google Secret Manager (e.g. gsm://secret/path?version=1)
//   - http: for HTTP resources (e.g. http://example.com/resource)
//   - https: for HTTPS resources (e.g. https://example.com/resource)
//   - s3: for Amazon S3 (e.g. s3://bucket/key)
//
// Users can also register custom backends for additional schemes using the [RegisterBackend] function.
type URIPath struct {
	// unexported variables
	backend  URIPathBackend      // backend which actually will be doing the work
	fragment string              // fragment portion of URI (after '#')
	host     string              // host portion of URI (after '//' and before next '/')
	password string              // password portion of URI (after ':' in username:password)
	path     string              // path portion of URI (after host and before '?')
	query    map[string][]string // query parameters (after '?')
	scheme   string              // scheme portion of URI (before '://')
	username string              // username portion of URI (before ':')
}

// RegisterBackend registers a backend for a specific scheme, allowing users to extend the functionality with
// custom backends.
//
// If a backend is already registered for the given scheme, it will be overwritten. Backends must be registered
// before calling [ParseURI] with a URI that uses the corresponding scheme.
func RegisterBackend(scheme string, newBackendFunc NewURIPathBackendFunc) {
	_registryMutex.Lock()
	defer _registryMutex.Unlock()
	_backendRegistry[strings.ToLower(scheme)] = newBackendFunc
}

// ParseURI parses a URI string and creates a URIPath with the appropriate backend.
//
// Typically options should be passed to the backend as query parameters in the URI.  However, there may be cases
// where this is not desirable so the backendOptions parameter can be used to pass options directly to the backend.
// Review the individual backend documentation for how options are used during construction.
//
// This function may return an error with any of the following codes:
//   - [InvalidParameter]: the URI is not valid
//   - [SchemeNotFound]: the URI scheme is not registered
func ParseURI(uri string, backendOptions ...map[string]any) (*URIPath, xerrors.Error) {
	// parse the URI
	parsedURL, err := url.Parse(uri)
	if err != nil {
		return nil, xerrors.Wrapf(InvalidParameter, err, "failed to parse URI '%s': %s", uri, err.Error()).
			WithAttr("uri", uri)
	}

	// validate the scheme exists in the registry and get the corresponding backend constructor
	scheme := strings.ToLower(parsedURL.Scheme)
	newBackendFunc, xerr := getBackend(scheme)
	if xerr != nil {
		return nil, xerr
	}

	// extract username and password
	var username, password string
	if parsedURL.User != nil {
		username = parsedURL.User.Username()
		password, _ = parsedURL.User.Password()
	}

	// construct the object
	u := &URIPath{
		scheme:   scheme,
		username: username,
		password: password,
		host:     parsedURL.Host,
		path:     parsedURL.Path,
		query:    parsedURL.Query(),
		fragment: parsedURL.Fragment,
	}
	u.backend, xerr = newBackendFunc(u, backendOptions...)
	if xerr != nil {
		return nil, xerr
	}
	return u, nil
}

// Backend returns the backend associated with this URIPath.
func (u *URIPath) Backend() URIPathBackend {
	return u.backend
}

// BackendAs returns the backend cast to the concrete type T.
//
// If the backend is not of type T, an error is returned.
func BackendAs[T URIPathBackend](u *URIPath) (T, xerrors.Error) {
	backend, ok := u.backend.(T)
	if !ok {
		var zero T
		return zero, xerrors.Newf(InvalidParameter, "backend is not of type %T", zero)
	}
	return backend, nil
}

// Delete removes the resource at the given path.
//
// The context and options passed to this function are passed directly to the backend. Be sure
// to review the documentation for the backend to understand how the context and options are used.
//
// This function may return an error with any of the codes from the [URIPathBackend.Delete] method.
func (u *URIPath) Delete(ctx context.Context, options map[string]any) xerrors.Error {
	return u.backend.Delete(ctx, options)
}

// Exists checks if the resource exists at the given path.
//
// The context and options passed to this function are passed directly to the backend. Be sure
// to review the documentation for the backend to understand how the context and options are used.
//
// This function may return an error with any of the codes from the [URIPathBackend.Exists] method.
func (u *URIPath) Exists(ctx context.Context, options map[string]any) (bool, xerrors.Error) {
	return u.backend.Exists(ctx, options)
}

// Fragment returns the URI fragment.
func (u *URIPath) Fragment() string {
	return u.fragment
}

// Get retrieves the content of the resource at the given path.
//
// The context and options passed to this function are passed directly to the backend. Be sure
// to review the documentation for the backend to understand how the context and options are used.
//
// This function may return an error with any of the codes from the [URIPathBackend.Get] method.
func (u *URIPath) Get(ctx context.Context, options map[string]any) ([]byte, xerrors.Error) {
	return u.backend.Get(ctx, options)
}

// Host returns the URI host.
func (u *URIPath) Host() string {
	return u.host
}

// List lists the resources at the given path.
//
// The context and options passed to this function are passed directly to the backend. Be sure
// to review the documentation for the backend to understand how the context and options are used.
//
// This function may return an error with any of the codes from the [URIPathBackend.List] method.
func (u *URIPath) List(ctx context.Context, recurse bool, options map[string]any) ([]string, xerrors.Error) {
	return u.backend.List(ctx, recurse, options)
}

// MarshalJSON marshals the [URIPath] value to JSON.
func (u URIPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

// MarshalText marshals the [URIPath] value to plain text.
func (u URIPath) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// Password returns the URI password.
func (u *URIPath) Password() string {
	return u.password
}

// Path returns the URI path.
func (u *URIPath) Path() string {
	return u.path
}

// Put stores content at the given path.
//
// The context and options passed to this function are passed directly to the backend. Be sure
// to review the documentation for the backend to understand how the context and options are used.
//
// This function may return an error with any of the codes from the [URIPathBackend.Put] method.
func (u *URIPath) Put(ctx context.Context, data []byte, options map[string]any) xerrors.Error {
	return u.backend.Put(ctx, data, options)
}

// Query returns the URI query parameters.
func (u *URIPath) Query() map[string][]string {
	return u.query
}

// Scheme returns the URI scheme (e.g., "s3", "file", "git").
func (u *URIPath) Scheme() string {
	return u.scheme
}

// String returns the full URI string.
func (u *URIPath) String() string {
	var builder strings.Builder
	builder.WriteString(u.scheme)
	builder.WriteString("://")

	if u.username != "" {
		builder.WriteString(u.username)
		if u.password != "" {
			builder.WriteString(":")
			builder.WriteString(u.password)
		}
		builder.WriteString("@")
	}

	builder.WriteString(u.host)
	builder.WriteString(u.path)

	if len(u.query) > 0 {
		builder.WriteString("?")
		first := true
		for key, values := range u.query {
			for _, value := range values {
				if !first {
					builder.WriteString("&")
				}
				builder.WriteString(url.QueryEscape(key))
				builder.WriteString("=")
				builder.WriteString(url.QueryEscape(value))
				first = false
			}
		}
	}

	if u.fragment != "" {
		builder.WriteString("#")
		builder.WriteString(u.fragment)
	}

	return builder.String()
}

// UnmarshalJSON parses the JSON data into a [URIPath] value.
func (u *URIPath) UnmarshalJSON(data []byte) error {
	var sval string
	if err := json.Unmarshal(data, &sval); err != nil {
		return err
	}
	val, err := ParseURI(sval)
	if err != nil {
		return err
	}
	u = val
	return nil
}

// UnmarshalText parses the text into a [URIPath] value.
func (u *URIPath) UnmarshalText(data []byte) error {
	val, err := ParseURI(string(data))
	if err != nil {
		return err
	}
	u = val
	return nil
}

// Username returns the URI username.
func (u *URIPath) Username() string {
	return u.username
}

// GetFnOptionValue is a helper function that returns the value of the option with the given key, first searching the
// function options list then the backend options. If the option is not found, the default value is returned.
//
// The value belonging to the matching key must be of the given type, otherwise it will be ignored.
func GetFnOptionValue[T any](defaultValue T, key string, backendOptions map[string]any, fnOptionsList ...map[string]any) T {
	for _, options := range fnOptionsList {
		if v, ok := options[key]; ok {
			if val, ok := v.(T); ok {
				return val
			}
		}
	}
	if v, ok := backendOptions[key]; ok {
		if val, ok := v.(T); ok {
			return val
		}
	}
	return defaultValue
}

// GetQueryOptionValue is a helper function that returns the value of the option with the given key, first searching the
// query parameters then the given options. If the option is not found, the default value is returned.
//
// The value belonging to the matching key must be of the given type, otherwise it will be ignored.
func GetQueryOptionValue[T any](defaultValue T, key string, queryParams url.Values, optionsList ...map[string]any) T {
	val := queryParams.Get(key)
	if val != "" {
		if v, err := convertString[T](val); err == nil {
			return v
		}
	}

	for _, options := range optionsList {
		if v, ok := options[key]; ok {
			if val, ok := v.(T); ok {
				return val
			}
		}
	}
	return defaultValue
}

// convertString converts a string to a type T using a type switch to handle common primitive types.
func convertString[T any](s string) (T, error) {
	var result T
	switch any(result).(type) {
	case int:
		v, err := strconv.Atoi(s)
		if err != nil {
			return result, err
		}
		// Conversion through any to satisfy the compiler
		result = any(v).(T)
	case int64:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return result, err
		}
		result = any(v).(T)
	case float64:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return result, err
		}
		result = any(v).(T)
	case string:
		result = any(s).(T)
	case bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return result, err
		}
		result = any(v).(T)
	default:
		// Handle unsupported types or panic
		return result, fmt.Errorf("unsupported type: %T", result)
	}
	return result, nil
}

// getBackend returns the backend factory function for a given scheme.
//
// This function may return an error with any of the following codes:
//   - [SchemeNotFound]: the scheme is not registered
func getBackend(scheme string) (NewURIPathBackendFunc, xerrors.Error) {
	_registryMutex.RLock()
	defer _registryMutex.RUnlock()

	if scheme == "" {
		scheme = FileScheme // default to file backend if no scheme is provided
	}
	if newBackendFunc, exists := _backendRegistry[strings.ToLower(scheme)]; exists {
		return newBackendFunc, nil
	}
	return nil, xerrors.Newf(SchemeNotFound, "%s: scheme not found", scheme)
}
