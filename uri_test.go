package uripath_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"go.innotegrity.dev/mod/xerrors"

	"go.innotegrity.dev/mod/uripath"
	urierrors "go.innotegrity.dev/mod/uripath/errors"
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	// Register a minimal s3 backend so URI tests can parse s3:// without importing backends/aws.
	_ = uripath.RegisterBackend(
		ctx,
		"s3",
		func(ctx context.Context, uri *uripath.URI, options ...uripath.BackendOption) (
			uripath.Backend, xerrors.Error,
		) {
			return &testBackend{}, nil
		},
		true,
	)

	os.Exit(m.Run())
}

func testContext(t *testing.T) context.Context {
	t.Helper()

	return context.Background()
}

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
	[]string, xerrors.Error,
) {
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

// otherBackend is a distinct [uripath.Backend] type used for type-mismatch assertions.
type otherBackend struct{ testBackend }

func TestRegisterBackend(t *testing.T) {
	ctx := testContext(t)
	// use a temporary scheme to avoid clashing with existing backends.
	xerr := uripath.RegisterBackend(
		ctx,
		"testscheme",
		func(ctx context.Context, uri *uripath.URI, options ...uripath.BackendOption) (
			uripath.Backend, xerrors.Error,
		) {
			return &testBackend{}, nil
		},
	)
	if xerr != nil {
		t.Fatalf("unexpected error registering backend: %v", xerr)
	}

	u, err := uripath.ParseURI(ctx, "testscheme://host/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if u.Scheme() != "testscheme" {
		t.Fatalf("expected scheme testscheme, got %q", u.Scheme())
	}

	// try to register an invalid scheme
	xerr = uripath.RegisterBackend(
		ctx,
		"invalid_scheme",
		func(ctx context.Context, uri *uripath.URI, options ...uripath.BackendOption) (
			uripath.Backend, xerrors.Error,
		) {
			return nil, nil
		},
	)
	if xerr == nil {
		t.Fatal("expected error from invalid backend registration")
	} else if xerr.Code() != urierrors.InvalidSchemeErrorCode {
		t.Fatalf("expected error code %d, got error code %d", urierrors.InvalidSchemeErrorCode, xerr.Code())
	}
}

func TestRegisterBackend_WithOverwrite(t *testing.T) {
	ctx := testContext(t)
	scheme := "overwrite"
	firstBackend := func(ctx context.Context, uri *uripath.URI, options ...uripath.BackendOption) (uripath.Backend, xerrors.Error) {
		return &testBackend{existsValue: false}, nil
	}
	secondBackend := func(ctx context.Context, uri *uripath.URI, options ...uripath.BackendOption) (uripath.Backend, xerrors.Error) {
		return &testBackend{existsValue: true}, nil
	}

	if xerr := uripath.RegisterBackend(ctx, scheme, firstBackend, true); xerr != nil {
		t.Fatalf("unexpected error registering first backend: %v", xerr)
	}

	xerr := uripath.RegisterBackend(ctx, scheme, secondBackend)
	if xerr == nil {
		t.Fatal("expected duplicate registration error without overwrite")
	} else if xerr.Code() != urierrors.SchemeExistsErrorCode {
		t.Fatalf("expected error code %d, got error code %d", urierrors.SchemeExistsErrorCode, xerr.Code())
	}

	err := uripath.RegisterBackend(ctx, scheme, secondBackend, true)
	if err != nil {
		t.Fatalf("unexpected error overwriting backend: %v", err)
	}
}

func TestBackendAs_CustomBackend(t *testing.T) {
	ctx := testContext(t)
	scheme := "custom"

	backend := func(ctx context.Context, uri *uripath.URI, options ...uripath.BackendOption) (uripath.Backend, xerrors.Error) {
		return &testBackend{existsValue: true}, nil
	}

	xerr := uripath.RegisterBackend(ctx, scheme, backend)
	if xerr != nil {
		t.Fatalf("unexpected error registering backend: %v", xerr)
	}

	u, err := uripath.ParseURI(ctx, "custom:///root/host/path")
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	b, castErr := uripath.BackendAs[*testBackend](ctx, u)
	if castErr != nil {
		t.Fatalf("unexpected cast error: %v", castErr)
	}

	if !b.existsValue {
		t.Fatal("expected overwritten backend to be in use")
	}
}

func TestURI_UnknownScheme(t *testing.T) {
	ctx := testContext(t)

	_, xerr := uripath.ParseURI(ctx, "nope://example/path")
	if xerr == nil {
		t.Fatal("expected error for unknown scheme")
	} else if xerr.Code() != urierrors.SchemeNotFoundErrorCode {
		t.Fatalf("expected error code %d, got error code %d", urierrors.SchemeNotFoundErrorCode, xerr.Code())
	}
}

func TestURI_BackendNewReturnsError(t *testing.T) {
	ctx := testContext(t)

	scheme := "failinit"
	if xerr := uripath.RegisterBackend(
		ctx,
		scheme,
		func(ctx context.Context, uri *uripath.URI, options ...uripath.BackendOption) (
			uripath.Backend, xerrors.Error,
		) {
			return nil, xerrors.Newf(urierrors.BackendInitErrorCode, "constructor failed")
		},
		true,
	); xerr != nil {
		t.Fatalf("unexpected register error: %v", xerr)
	}

	_, xerr := uripath.ParseURI(ctx, scheme+"://host/path")
	if xerr == nil {
		t.Fatal("expected error when backend constructor fails")
	}

	if xerr.Code() != urierrors.BackendInitErrorCode {
		t.Fatalf("expected error code %d, got %d", urierrors.BackendInitErrorCode, xerr.Code())
	}
}

func TestURI_GetBackendInvalidScheme(t *testing.T) {
	ctx := testContext(t)
	t.Run("scheme format invalid", func(t *testing.T) {
		// getBackend runs validateScheme; a scheme must start with a letter.
		_, xerr := uripath.ParseURI(ctx, "invalid=scheme://host/path")
		if xerr == nil {
			t.Fatal("expected error for invalid scheme format")
		}

		if xerr.Code() != urierrors.InvalidParameterErrorCode {
			t.Fatalf("expected error code %d, got %d", urierrors.InvalidParameterErrorCode, xerr.Code())
		}
	})

	t.Run("scheme not registered", func(t *testing.T) {
		_, xerr := uripath.ParseURI(ctx, "notregistered://host/path")
		if xerr == nil {
			t.Fatal("expected error for unregistered scheme")
		}

		if xerr.Code() != urierrors.SchemeNotFoundErrorCode {
			t.Fatalf("expected error code %d, got %d", urierrors.SchemeNotFoundErrorCode, xerr.Code())
		}
	})
}

func TestURI_MarshalJSONAndText(t *testing.T) {
	ctx := testContext(t)
	rawURI := "s3://aws_access_key_id:aws_secret_access_key@some.bucket.url/path/to/file.txt?x=1&y=2#frag"

	u, xerr := uripath.ParseURI(ctx, rawURI)
	if xerr != nil {
		t.Fatalf("unexpected error: %v", xerr)
	}

	if u.String() != rawURI {
		t.Fatalf("String should preserve credentials, got %q", u.String())
	}

	// "aws_access_key_id" -> first/last rune kept, 15 '*' in between
	const wantRedacted = "s3://a***************d:********@some.bucket.url/path/to/file.txt?x=1&y=2#frag"

	jsonBytes, err := u.MarshalJSON()
	if err != nil {
		t.Fatalf("unexpected error from MarshalJSON: %v", err)
	}

	var decoded string
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("unexpected json unmarshal error: %v", err)
	}

	if decoded != wantRedacted {
		t.Fatalf("expected marshaled JSON string %q, got %q", wantRedacted, decoded)
	}

	textBytes, err := u.MarshalText()
	if err != nil {
		t.Fatalf("unexpected error from MarshalText: %v", err)
	}

	if string(textBytes) != wantRedacted {
		t.Fatalf("expected text value %q, got %q", wantRedacted, string(textBytes))
	}

	var roundTrip uripath.URI
	if err := json.Unmarshal(jsonBytes, &roundTrip); err != nil {
		t.Fatalf("unmarshal marshaled JSON: %v", err)
	}

	if roundTrip.Username == u.Username {
		t.Fatal("expected redacted username after JSON round-trip, got original username")
	}

	if roundTrip.Password != "********" {
		t.Fatalf("expected placeholder password after round-trip, got %q", roundTrip.Password)
	}
}

func TestURI_SafeString(t *testing.T) {
	ctx := testContext(t)

	u, xerr := uripath.ParseURI(ctx, "s3://key:secret@host/p")
	if xerr != nil {
		t.Fatalf("parse: %v", xerr)
	}
	// "key" -> k*y (3 runes)
	const want = "s3://k*y:********@host/p"
	if got := u.SafeString(); got != want {
		t.Fatalf("SafeString: got %q want %q", got, want)
	}

	jsonBytes, err := u.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	var fromJSON string
	if err := json.Unmarshal(jsonBytes, &fromJSON); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if fromJSON != u.SafeString() {
		t.Fatalf("MarshalJSON string %q != SafeString %q", fromJSON, u.SafeString())
	}

	var nilURI *uripath.URI
	if nilURI.SafeString() != "" {
		t.Fatalf("nil SafeString should be empty, got %q", nilURI.SafeString())
	}
}

func TestURI_MarshalJSONAndText_NoUserInfo(t *testing.T) {
	ctx := testContext(t)
	raw := "s3://some.bucket.url/path/to/file.txt?x=1&y=2#frag"

	u, xerr := uripath.ParseURI(ctx, raw)
	if xerr != nil {
		t.Fatalf("unexpected error: %v", xerr)
	}

	if u.String() != raw {
		t.Fatalf("String: %q", u.String())
	}

	jb, err := u.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	var s string
	if err := json.Unmarshal(jb, &s); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if s != raw {
		t.Fatalf("expected no redaction change without userinfo, got %q", s)
	}
}

func TestURI_BackendAs(t *testing.T) {
	ctx := testContext(t)
	stub := &testBackend{existsValue: true}
	u := &uripath.URI{Backend: stub}

	asStub, err := uripath.BackendAs[*testBackend](ctx, u)
	if err != nil {
		t.Fatalf("expected no error from BackendAs[*testBackend], got %v", err)
	}

	if asStub != stub {
		t.Fatal("expected the same backend from BackendAs")
	}

	_, err = uripath.BackendAs[*otherBackend](ctx, u)
	if err == nil {
		t.Fatal("expected error from BackendAs[*otherBackend] when backend type mismatch")
	}
}

func TestURI_DelegatedMethods(t *testing.T) {
	ctx := testContext(t)
	stub := &testBackend{existsValue: true}
	u := &uripath.URI{Backend: stub}

	if _, err := u.Exists(ctx); err != nil {
		t.Fatalf("unexpected Exists error: %v", err)
	}

	if !stub.existsCalled {
		t.Fatal("Expected Exists to be called on backend")
	}

	if _, err := u.Get(ctx); err != nil {
		t.Fatalf("unexpected Get error: %v", err)
	}

	if !stub.getCalled {
		t.Fatal("Expected Get to be called on backend")
	}

	if _, err := u.List(ctx, true); err != nil {
		t.Fatalf("unexpected List error: %v", err)
	}

	if !stub.listCalled {
		t.Fatal("Expected List to be called on backend")
	}

	err := u.Put(ctx, []byte("abc"))
	if err != nil {
		t.Fatalf("unexpected Put error: %v", err)
	}

	if !stub.putCalled {
		t.Fatal("Expected Put to be called on backend")
	}

	err = u.Delete(ctx)
	if err != nil {
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
	jsonData := fmt.Appendf(nil, `"%s"`, uri)

	var u1 uripath.URI

	err := json.Unmarshal(jsonData, &u1)
	if err != nil {
		t.Fatalf("unexpected error from UnmarshalJSON: %v", err)
	}

	if u1.Scheme() != "s3" {
		t.Fatalf("expected scheme %q, got %q", "s3", u1.Scheme())
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

	if u3.Scheme() != "s3" {
		t.Fatalf("expected scheme %q, got %q", "s3", u3.Scheme())
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
	ctx := testContext(t)

	_, xerr := uripath.ParseURI(ctx, "foo://user:pass%x@bar")
	if xerr == nil {
		t.Fatal("expected error from ParseURI with invalid URI")
	} else if xerr.Code() != urierrors.InvalidParameterErrorCode {
		t.Fatalf("expected error code %d, got error code %d", urierrors.InvalidParameterErrorCode, xerr.Code())
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
	ctx := testContext(t)
	uri := "s3://aws_access_key_id:aws_secret_access_key@some.bucket.url/path/to/file.txt?x=1&y=2#frag"

	u, err := uripath.ParseURI(ctx, uri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := u.String()
	if u.String() != uri {
		t.Fatalf("expected output %q, got %q", uri, output)
	}
}
