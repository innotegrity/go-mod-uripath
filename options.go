package uripath

import (
	"net/url"
)

// BackendOption is a function that can be used to pass custom options to a backend.
type BackendOption func(URIPathBackend)

// GetFnOptionValue is a helper function that returns the value of the option with the given key, first searching the
// function options list then the backend options. If the option is not found, the default value is returned.
//
// The value belonging to the matching key must be of the given type, otherwise it will be ignored.
func GetFnOptionValue[T any](defaultValue T, key string, backendOptions map[string]any, fnOptions ...BackendOption) T {
	// create a fake base backend to apply the options to
	base := InitBackendBase(nil, fnOptions...)
	options := base.Options()
	if v, ok := options[key]; ok {
		if val, ok := v.(T); ok {
			return val
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
func GetQueryOptionValue[T any](defaultValue T, key string, queryParams url.Values, fnOptions ...BackendOption) T {
	if queryParams != nil {
		val := queryParams.Get(key)
		if val != "" {
			if v, err := convertString[T](val); err == nil {
				return v
			}
		}
	}

	// create a fake base backend to apply the options to
	base := InitBackendBase(nil, fnOptions...)
	options := base.Options()
	if v, ok := options[key]; ok {
		if val, ok := v.(T); ok {
			return val
		}
	}
	return defaultValue
}

// WithBackendOption returns a [BackendOption] that sets the option with the given key to the given value.
func WithBackendOption(key string, value any) BackendOption {
	return func(b URIPathBackend) {
		b.SetOption(key, value)
	}
}
