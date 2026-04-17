package uripath

import (
	"net/url"
	"testing"
)

func TestWithBackendOptions(t *testing.T) {
	opts := WithBackendOptions(map[string]any{
		"count":   42,
		"label":   "hello",
		"enabled": true,
	})

	if len(opts) != 3 {
		t.Fatalf("expected 3 backend options, got %d", len(opts))
	}

	if got := GetFnOptionValue(0, "count", nil, opts...); got != 42 {
		t.Fatalf("expected count 42, got %v", got)
	}
	if got := GetFnOptionValue("", "label", nil, opts...); got != "hello" {
		t.Fatalf("expected label hello, got %q", got)
	}
	if got := GetFnOptionValue(false, "enabled", nil, opts...); got != true {
		t.Fatalf("expected enabled true, got %v", got)
	}
}

func TestGetFnOptionValue_FromFnOptions(t *testing.T) {
	t.Run("int from WithBackendOption", func(t *testing.T) {
		got := GetFnOptionValue(100, "n", nil, WithBackendOption("n", 7))
		if got != 7 {
			t.Fatalf("expected 7, got %v", got)
		}
	})

	t.Run("string from WithBackendOption", func(t *testing.T) {
		got := GetFnOptionValue("default", "s", nil, WithBackendOption("s", "from-option"))
		if got != "from-option" {
			t.Fatalf("expected from-option, got %q", got)
		}
	})

	t.Run("bool from WithBackendOption", func(t *testing.T) {
		got := GetFnOptionValue(false, "flag", nil, WithBackendOption("flag", true))
		if !got {
			t.Fatalf("expected true, got %v", got)
		}
	})

	t.Run("fn options take precedence over backend map", func(t *testing.T) {
		backendOpts := map[string]any{"n": 1}
		got := GetFnOptionValue(0, "n", backendOpts, WithBackendOption("n", 99))
		if got != 99 {
			t.Fatalf("expected fn option 99, got %v", got)
		}
	})
}

func TestGetFnOptionValue_BackendOptionUnexpectedType(t *testing.T) {
	backendOptions := map[string]any{"max": "should-be-int"}
	value := GetFnOptionValue(100, "max", backendOptions)
	if value != 100 {
		t.Fatalf("expected default 100 because backend type mismatch, got %v", value)
	}

	value2 := GetFnOptionValue("", "max", backendOptions)
	if value2 != "should-be-int" {
		t.Fatalf("expected 'should-be-int', got %v", value2)
	}
}

func TestGetQueryOptionValue_BackendOptionUnexpectedType(t *testing.T) {
	query := url.Values{}
	value := GetQueryOptionValue(5, "limit", query, WithBackendOption("limit", "not-int"))
	if value != 5 {
		t.Fatalf("expected default 5 because backend type mismatch, got %v", value)
	}

	value2 := GetQueryOptionValue("", "limit", query, WithBackendOption("limit", "none"))
	if value2 != "none" {
		t.Fatalf("expected '10', got %v", value2)
	}
}

func TestGetQueryOptionValue_TypeConversions(t *testing.T) {
	query := url.Values{}

	// Test int conversion
	query.Set("intkey", "1234")
	intVal := GetQueryOptionValue(int(0), "intkey", query)
	if intVal != 1234 {
		t.Fatalf("expected 1234, got %v", intVal)
	}

	// Test int64 conversion
	query.Set("int64key", "123456789")
	int64Val := GetQueryOptionValue(int64(0), "int64key", query)
	if int64Val != 123456789 {
		t.Fatalf("expected 123456789, got %v", int64Val)
	}

	// Test float64 conversion
	query.Set("floatkey", "3.14")
	floatVal := GetQueryOptionValue(0.0, "floatkey", query)
	if floatVal != 3.14 {
		t.Fatalf("expected 3.14, got %v", floatVal)
	}

	// Test bool conversion
	query.Set("boolkey", "true")
	boolVal := GetQueryOptionValue(false, "boolkey", query)
	if boolVal != true {
		t.Fatalf("expected true, got %v", boolVal)
	}

	// Test string conversion (should just return the string)
	query.Set("strkey", "hello")
	strVal := GetQueryOptionValue("", "strkey", query)
	if strVal != "hello" {
		t.Fatalf("expected 'hello', got %v", strVal)
	}

	// Test invalid int conversion falls back to default
	query.Set("badint", "notanumber")
	badIntVal := GetQueryOptionValue(99, "badint", query)
	if badIntVal != 99 {
		t.Fatalf("expected default 99 for invalid int, got %v", badIntVal)
	}

	// Test invalid int64 conversion
	query.Set("badint64", "abc")
	badInt64Val := GetQueryOptionValue(int64(88), "badint64", query)
	if badInt64Val != 88 {
		t.Fatalf("expected default 88 for invalid int64, got %v", badInt64Val)
	}

	// Test invalid float64 conversion
	query.Set("badfloat", "notafloat")
	badFloatVal := GetQueryOptionValue(7.7, "badfloat", query)
	if badFloatVal != 7.7 {
		t.Fatalf("expected default 7.7 for invalid float64, got %v", badFloatVal)
	}

	// Test invalid bool conversion
	query.Set("badbool", "notabool")
	badBoolVal := GetQueryOptionValue(true, "badbool", query)
	if badBoolVal != true {
		t.Fatalf("expected default true for invalid bool, got %v", badBoolVal)
	}
}

func TestGetQueryOptionValue_ComplexTypeConversions(t *testing.T) {
	query := url.Values{}
	query.Set("complexkey", `{"a": 1, "b": 2}`)

	defaultVal := map[string]int{"a": 1, "b": 2}
	complexVal := GetQueryOptionValue(defaultVal, "complexkey", query)
	if complexVal == nil {
		t.Fatalf("expected non-nil complex value, got nil")
	}
}
