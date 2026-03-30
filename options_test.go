package uripath

import (
	"net/url"
	"testing"
)

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
