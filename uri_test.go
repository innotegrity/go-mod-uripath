package uripath_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"go.innotegrity.dev/mod/xerrors"

	"go.innotegrity.dev/mod/uripath"
	"go.innotegrity.dev/mod/uripath/backends/aws"
)

// testBackend is a minimal [uripath.Backend] stub used in tests.
type testBackend struct {
	deleteCalled bool
	existsCalled bool
	existsValue  bool
	getCalled    bool
	listCalled   bool
	putCalled    bool
}

// Delete implements [uripath.Backend].
func (b *testBackend) Delete(ctx context.Context, options ...uripath.BackendOption) xerrors.Error {
	b.deleteCalled = true
	return nil
}

// Exists implements [uripath.Backend].
func (b *testBackend) Exists(ctx context.Context, options ...uripath.BackendOption) (bool, xerrors.Error) {
	b.existsCalled = true
	return b.existsValue, nil
}

// Get implements [uripath.Backend].
func (b *testBackend) Get(ctx context.Context, options ...uripath.BackendOption) ([]byte, xerrors.Error) {
	b.getCalled = true
	return []byte("data"), nil
}

// List implements [uripath.Backend].
func (b *testBackend) List(ctx context.Context, recurse bool, options ...uripath.BackendOption) (
	[]string, xerrors.Error) {

	b.listCalled = true
	return []string{"a", "b"}, nil
}

// Options implements [uripath.Backend].
func (b *testBackend) Options() map[string]any {
	return map[string]any{}
}

// Put implements [uripath.Backend].
func (b *testBackend) Put(ctx context.Context, data []byte, options ...uripath.BackendOption) xerrors.Error {
	b.putCalled = true
	return nil
}

// RemoveAllOptions implements [uripath.Backend].
func (b *testBackend) RemoveAllOptions() {
}

// RemoveOption implements [uripath.Backend].
func (b *testBackend) RemoveOption(key string) uripath.Backend {
	return b
}

// ReplaceOptions implements [uripath.Backend].
func (b *testBackend) ReplaceOptions(options map[string]any) {
}

// SetOption implements [uripath.Backend].
func (b *testBackend) SetOption(key string, value any) uripath.Backend {
	return b
}

// URI implements [uripath.Backend].
func (b *testBackend) URI() *uripath.URI {
	return nil
}

func TestRegisterBackend(t *testing.T) {
	// use a temporary scheme to avoid clashing with existing backends.
	xerr := uripath.RegisterBackend("testscheme", func(uri *uripath.URI, options ...uripath.BackendOption) (
		uripath.Backend, xerrors.Error) {
		return &testBackend{}, nil
	})
	if xerr != nil {
		t.Fatalf("unexpected error registering backend: %v", xerr)
	}

	u, err := uripath.ParseURI("testscheme://host/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Scheme() != "testscheme" {
		t.Fatalf("expected scheme testscheme, got %q", u.Scheme())
	}

	// try to register an invalid scheme
	xerr = uripath.RegisterBackend("invalid_scheme", func(uri *uripath.URI, options ...uripath.BackendOption) (
		uripath.Backend, xerrors.Error) {
		return nil, nil
	})
	if xerr == nil {
		t.Fatal("expected error from invalid backend registration")
	} else if xerr.Code() != uripath.InvalidParameter {
		t.Fatalf("expected error code %d, got error code %d", uripath.InvalidParameter, xerr.Code())
	}
}

func TestRegisterBackend_WithOverwrite(t *testing.T) {
	scheme := "overwrite"
	firstBackend := func(uri *uripath.URI, options ...uripath.BackendOption) (uripath.Backend, xerrors.Error) {
		return &testBackend{existsValue: false}, nil
	}
	secondBackend := func(uri *uripath.URI, options ...uripath.BackendOption) (uripath.Backend, xerrors.Error) {
		return &testBackend{existsValue: true}, nil
	}

	if xerr := uripath.RegisterBackend(scheme, firstBackend, true); xerr != nil {
		t.Fatalf("unexpected error registering first backend: %v", xerr)
	}
	xerr := uripath.RegisterBackend(scheme, secondBackend)
	if xerr == nil {
		t.Fatal("expected duplicate registration error without overwrite")
	} else if xerr.Code() != uripath.SchemeExists {
		t.Fatalf("expected error code %d, got error code %d", uripath.SchemeExists, xerr.Code())
	}
	if err := uripath.RegisterBackend(scheme, secondBackend, true); err != nil {
		t.Fatalf("unexpected error overwriting backend: %v", err)
	}
}

func TestBackendAs_CustomBackend(t *testing.T) {
	scheme := "custom"
	backend := func(uri *uripath.URI, options ...uripath.BackendOption) (uripath.Backend, xerrors.Error) {
		return &testBackend{existsValue: true}, nil
	}
	if xerr := uripath.RegisterBackend(scheme, backend); xerr != nil {
		t.Fatalf("unexpected error registering backend: %v", xerr)
	}

	u, err := uripath.ParseURI("custom:///root/host/path")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	b, castErr := uripath.BackendAs[*testBackend](u)
	if castErr != nil {
		t.Fatalf("unexpected cast error: %v", castErr)
	}
	if !b.existsValue {
		t.Fatal("expected overwritten backend to be in use")
	}
}

func TestURI_UnknownScheme(t *testing.T) {
	_, xerr := uripath.ParseURI("nope://example/path")
	if xerr == nil {
		t.Fatal("expected error for unknown scheme")
	} else if xerr.Code() != uripath.SchemeNotFound {
		t.Fatalf("expected error code %d, got error code %d", uripath.SchemeNotFound, xerr.Code())
	}
}

func TestURI_BackendNewReturnsError(t *testing.T) {
	scheme := "failinit"
	if xerr := uripath.RegisterBackend(scheme, func(uri *uripath.URI, options ...uripath.BackendOption) (
		uripath.Backend, xerrors.Error) {
		return nil, xerrors.Newf(uripath.BackendInitError, "constructor failed")
	}, true); xerr != nil {
		t.Fatalf("unexpected register error: %v", xerr)
	}

	_, xerr := uripath.ParseURI(scheme + "://host/path")
	if xerr == nil {
		t.Fatal("expected error when backend constructor fails")
	}
	if xerr.Code() != uripath.BackendInitError {
		t.Fatalf("expected error code %d, got %d", uripath.BackendInitError, xerr.Code())
	}
}

func TestURI_GetBackendInvalidScheme(t *testing.T) {
	t.Run("scheme format invalid", func(t *testing.T) {
		// getBackend runs validateScheme; a scheme must start with a letter.
		_, xerr := uripath.ParseURI("invalid=scheme://host/path")
		if xerr == nil {
			t.Fatal("expected error for invalid scheme format")
		}
		if xerr.Code() != uripath.InvalidParameter {
			t.Fatalf("expected error code %d, got %d", uripath.InvalidParameter, xerr.Code())
		}
	})

	t.Run("scheme not registered", func(t *testing.T) {
		_, xerr := uripath.ParseURI("notregistered://host/path")
		if xerr == nil {
			t.Fatal("expected error for unregistered scheme")
		}
		if xerr.Code() != uripath.SchemeNotFound {
			t.Fatalf("expected error code %d, got %d", uripath.SchemeNotFound, xerr.Code())
		}
	})
}

func TestURI_MarshalJSONAndText(t *testing.T) {
	uri := "s3://aws_access_key_id:aws_secret_access_key@some.bucket.url/path/to/file.txt?x=1&y=2#frag"

	u, xerr := uripath.ParseURI(uri)
	if xerr != nil {
		t.Fatalf("unexpected error: %v", xerr)
	}

	jsonBytes, err := u.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error from MarshalJSON: %v", err)
	}

	var decoded string
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("unexpected json unmarshal error: %v", err)
	}
	if decoded != u.String() {
		t.Fatalf("expected marshaled value %q, got %q", u.String(), decoded)
	}

	textBytes, err := u.MarshalText()
	if err != nil {
		t.Fatalf("unexpected error from MarshalText: %v", err)
	}
	if string(textBytes) != u.String() {
		t.Fatalf("expected text value %q, got %q", u.String(), string(textBytes))
	}
}

func TestURI_BackendAs(t *testing.T) {
	stub := &testBackend{existsValue: true}
	u := &uripath.URI{Backend: stub}

	asStub, err := uripath.BackendAs[*testBackend](u)
	if err != nil {
		t.Fatalf("expected no error from BackendAs[*testBackend], got %v", err)
	}
	if asStub != stub {
		t.Fatal("expected the same backend from BackendAs")
	}

	_, err = uripath.BackendAs[*aws.S3Backend](u)
	if err == nil {
		t.Fatal("expected error from BackendAs[*aws.S3Backend] when backend type mismatch")
	}
}

func TestURI_DelegatedMethods(t *testing.T) {
	stub := &testBackend{existsValue: true}
	u := &uripath.URI{Backend: stub}

	if _, err := u.Exists(context.Background()); err != nil {
		t.Fatalf("unexpected Exists error: %v", err)
	}
	if !stub.existsCalled {
		t.Fatal("Expected Exists to be called on backend")
	}

	if _, err := u.Get(context.Background()); err != nil {
		t.Fatalf("unexpected Get error: %v", err)
	}
	if !stub.getCalled {
		t.Fatal("Expected Get to be called on backend")
	}

	if _, err := u.List(context.Background(), true); err != nil {
		t.Fatalf("unexpected List error: %v", err)
	}
	if !stub.listCalled {
		t.Fatal("Expected List to be called on backend")
	}

	if err := u.Put(context.Background(), []byte("abc")); err != nil {
		t.Fatalf("unexpected Put error: %v", err)
	}
	if !stub.putCalled {
		t.Fatal("Expected Put to be called on backend")
	}

	if err := u.Delete(context.Background()); err != nil {
		t.Fatalf("unexpected Delete error: %v", err)
	}
	if !stub.deleteCalled {
		t.Fatal("Expected Delete to be called on backend")
	}

	if u.Backend != stub {
		t.Fatal("Expected Backend() to return underlying backend")
	}
}

func TestURI_UnmarshalJSONAndText(t *testing.T) {
	uri := "s3://aws_access_key_id:aws_secret_access_key@some.bucket.url/path/to/file.txt?x=1&y=2#frag"
	expectedPath := "/path/to/file.txt"
	expectedQuery := map[string]string{
		"x": "1",
		"y": "2",
	}
	expectedFragment := "frag"

	// Test UnmarshalJSON success
	jsonData := []byte(fmt.Sprintf(`"%s"`, uri))
	var u1 uripath.URI
	err := json.Unmarshal(jsonData, &u1)
	if err != nil {
		t.Fatalf("unexpected error from UnmarshalJSON: %v", err)
	}

	if u1.Scheme() != aws.S3Scheme {
		t.Fatalf("expected scheme %q, got %q", aws.S3Scheme, u1.Scheme())
	}
	if u1.Path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, u1.Path)
	}
	if u1.Fragment != expectedFragment {
		t.Fatalf("expected fragment %q, got %q", expectedFragment, u1.Fragment)
	}
	if q := u1.Query; q.Get("x") != expectedQuery["x"] || q.Get("y") != expectedQuery["y"] {
		t.Fatalf("unexpected query map: %v", q)
	}

	// Test UnmarshalJSON with invalid JSON
	invalidJSONData := []byte{'{', 7, '}'}
	var u2 uripath.URI
	err = u2.UnmarshalJSON(invalidJSONData)
	if err == nil {
		t.Fatal("expected error from UnmarshalJSON with invalid JSON")
	}

	// Test UnmarshalText success
	textData := []byte(uri)
	var u3 uripath.URI
	err = u3.UnmarshalText(textData)
	if err != nil {
		t.Fatalf("unexpected error from UnmarshalText: %v", err)
	}

	if u3.Scheme() != aws.S3Scheme {
		t.Fatalf("expected scheme %q, got %q", aws.S3Scheme, u3.Scheme())
	}
	if u3.Path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, u3.Path)
	}
	if u3.Fragment != expectedFragment {
		t.Fatalf("expected fragment %q, got %q", expectedFragment, u3.Fragment)
	}
	if q := u3.Query; q.Get("x") != expectedQuery["x"] || q.Get("y") != expectedQuery["y"] {
		t.Fatalf("unexpected query map: %v", q)
	}
}

func TestURI_InvalidURI(t *testing.T) {
	_, xerr := uripath.ParseURI("foo://user:pass%x@bar")
	if xerr == nil {
		t.Fatal("expected error from ParseURI with invalid URI")
	} else if xerr.Code() != uripath.InvalidParameter {
		t.Fatalf("expected error code %d, got error code %d", uripath.InvalidParameter, xerr.Code())
	}

	u := &uripath.URI{}
	err := u.UnmarshalJSON([]byte(`"foo://user:pass%x@bar"`))
	if err == nil {
		t.Fatal("expected error from UnmarshalJSON with invalid URI")
	}

	err = u.UnmarshalText([]byte("foo://user:pass%x@bar"))
	if err == nil {
		t.Fatal("expected error from UnmarshalText with invalid URI")
	}
}

func TestURI_String(t *testing.T) {
	uri := "s3://aws_access_key_id:aws_secret_access_key@some.bucket.url/path/to/file.txt?x=1&y=2#frag"
	u, err := uripath.ParseURI(uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := u.String()
	if u.String() != uri {
		t.Fatalf("expected output %q, got %q", uri, output)
	}
}
