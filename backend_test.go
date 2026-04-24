package uripath_test

import (
	"context"
	"testing"

	"go.innotegrity.dev/mod/uripath"
)

func TestBackendBase_Delete(t *testing.T) {
	bb := uripath.InitBackendBase(nil)

	err := bb.Delete(context.Background())
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestBackendBase_Exists(t *testing.T) {
	bb := uripath.InitBackendBase(nil)

	exists, err := bb.Exists(context.Background())
	if exists != false {
		t.Errorf("Expected false, got %v", exists)
	}

	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestBackendBase_Get(t *testing.T) {
	bb := uripath.InitBackendBase(nil)

	_, err := bb.Get(context.Background())
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestBackendBase_List(t *testing.T) {
	bb := uripath.InitBackendBase(nil)

	_, err := bb.List(context.Background(), false)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}

	_, err = bb.List(context.Background(), true)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestBackendBase_Put(t *testing.T) {
	bb := uripath.InitBackendBase(nil)

	err := bb.Put(context.Background(), []byte{})
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

func TestBackendBase_Options(t *testing.T) {
	bb := uripath.InitBackendBase(nil, uripath.WithBackendOption("test", "value"))

	options := bb.Options()
	if options == nil {
		t.Errorf("Expected non-nil options, got nil")
	}

	if options["test"] != "value" {
		t.Errorf("Expected option value to be 'value', got %v", options["test"])
	}
}

func TestBackendBase_RemoveAllOptions(t *testing.T) {
	bb := uripath.InitBackendBase(nil, uripath.WithBackendOption("test", "value"))
	bb.RemoveAllOptions()

	options := bb.Options()
	if len(options) != 0 {
		t.Errorf("Expected empty options, got %v", options)
	}
}

func TestBackendBase_RemoveOption(t *testing.T) {
	bb := uripath.InitBackendBase(nil, uripath.WithBackendOption("test", "value"))
	bb.RemoveOption("test1")

	options := bb.Options()
	if len(options) != 1 {
		t.Errorf("Expected 1 options, got %v", options)
	}

	if options["test"] != "value" {
		t.Errorf("Expected option value to be 'value', got %v", options["test"])
	}

	bb.RemoveOption("test")

	options = bb.Options()
	if len(options) != 0 {
		t.Errorf("Expected empty options, got %v", options)
	}
}

func TestBackendBase_ReplaceOptions(t *testing.T) {
	bb := uripath.InitBackendBase(nil, uripath.WithBackendOption("test", "value"))
	bb.ReplaceOptions(map[string]any{"test": "value1"})

	options := bb.Options()
	if len(options) != 1 {
		t.Errorf("Expected 1 options, got %v", options)
	}

	if options["test"] != "value1" {
		t.Errorf("Expected option value to be 'value1', got %v", options["test"])
	}
}

func TestBackendBase_SetOption(t *testing.T) {
	bb := uripath.InitBackendBase(nil)
	bb.SetOption("test", "value")

	options := bb.Options()
	if len(options) != 1 {
		t.Errorf("Expected 1 options, got %v", options)
	}

	if options["test"] != "value" {
		t.Errorf("Expected option value to be 'value', got %v", options["test"])
	}
}

func TestBackendBase_URI(t *testing.T) {
	bb := uripath.InitBackendBase(nil)
	if bb.URI() != nil {
		t.Errorf("Expected nil URI, got %v", bb.URI())
	}
}
