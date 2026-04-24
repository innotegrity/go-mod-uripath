package uripath

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

var (
	// errUnsupportedStringConversion is returned when [convertString] is called with a type parameter T
	// that is not one of the supported primitive mappings.
	errUnsupportedStringConversion = errors.New("unsupported type for convertString")
	errConvertStringToInt          = errors.New("failed to convert string to int")
	errConvertStringToInt64        = errors.New("failed to convert string to int64")
	errConvertStringToFloat64      = errors.New("failed to convert string to float64")
	errConvertStringToBool         = errors.New("failed to convert string to bool")
)

// BackendOption is a function that can be used to pass custom options to a backend.
type BackendOption func(Backend)

// GetFnOptionValue is a helper function that returns the value of the option with the given key, first searching the
// function options list then the backend options. If the option is not found, the default value is returned.
//
// The value belonging to the matching key must be of the given type, otherwise it will be ignored.
//
// //nolint:ireturn // want to be able to return generic type.
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
//
//nolint:ireturn // want to be able to return generic type.
func GetQueryOptionValue[T any](defaultValue T, key string, queryParams url.Values, fnOptions ...BackendOption) T {
	if queryParams != nil {
		val := queryParams.Get(key)
		if val != "" {
			v, err := convertString[T](val)
			if err == nil {
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
	return func(b Backend) {
		b.SetOption(key, value)
	}
}

// WithBackendOptions returns a list of [BackendOption]s that set the options in the given map.
func WithBackendOptions(options map[string]any) []BackendOption {
	result := make([]BackendOption, 0, len(options))
	for k, v := range options {
		result = append(result, func(b Backend) {
			b.SetOption(k, v)
		})
	}

	return result
}

// convertString converts a string to a type T using a type switch to handle common primitive types.
//
//nolint:ireturn // want to be able to return generic type.
func convertString[T any](str string) (T, error) {
	var result T
	switch any(result).(type) {
	case int:
		v, err := strconv.Atoi(str)
		if err != nil {
			return result, errors.Join(errConvertStringToInt, err)
		}

		result, _ = any(v).(T)
	case int64:
		v, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return result, errors.Join(errConvertStringToInt64, err)
		}

		result, _ = any(v).(T)
	case float64:
		v, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return result, errors.Join(errConvertStringToFloat64, err)
		}

		result, _ = any(v).(T)
	case string:
		result, _ = any(str).(T)
	case bool:
		v, err := strconv.ParseBool(str)
		if err != nil {
			return result, errors.Join(errConvertStringToBool, err)
		}

		result, _ = any(v).(T)
	default:
		return result, fmt.Errorf("%T: %w", result, errUnsupportedStringConversion)
	}

	return result, nil
}
