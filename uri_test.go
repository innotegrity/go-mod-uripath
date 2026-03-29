package uripath

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"runtime"
	"testing"

	"go.innotegrity.dev/mod/xerrors"
)

type testBackend struct {
	deleteCalled bool
	existsCalled bool
	existsValue  bool
	getCalled    bool
	listCalled   bool
	putCalled    bool
}

func (b *testBackend) Delete(ctx context.Context, options ...BackendOption) xerrors.Error {
	b.deleteCalled = true
	return nil
}

func (b *testBackend) Exists(ctx context.Context, options ...BackendOption) (bool, xerrors.Error) {
	b.existsCalled = true
	return b.existsValue, nil
}

func (b *testBackend) Get(ctx context.Context, options ...BackendOption) ([]byte, xerrors.Error) {
	b.getCalled = true
	return []byte("data"), nil
}

func (b *testBackend) List(ctx context.Context, recurse bool, options ...BackendOption) ([]string, xerrors.Error) {
	b.listCalled = true
	return []string{"a", "b"}, nil
}

func (b *testBackend) Options() map[string]any {
	return map[string]any{}
}

func (b *testBackend) Put(ctx context.Context, data []byte, options ...BackendOption) xerrors.Error {
	b.putCalled = true
	return nil
}

func (b *testBackend) RemoveAllOptions() {
}

func (b *testBackend) RemoveOption(key string) URIPathBackend {
	return b
}

func (b *testBackend) ReplaceOptions(options map[string]any) {
}

func (b *testBackend) SetOption(key string, value any) URIPathBackend {
	return b
}

func (b *testBackend) URIPath() *URIPath {
	return nil
}

func TestParseURI_UnknownScheme(t *testing.T) {
	_, err := ParseURI("nope://example/path")
	if err == nil {
		t.Fatal("expected error for unknown scheme")
	}
}

func TestParseURI_InvalidURI(t *testing.T) {
	_, err := ParseURI("://")
	if err == nil {
		t.Fatal("expected error for invalid URI")
	}
}

func TestParseURI_File(t *testing.T) {
	u, err := ParseURI("file:///tmp/testfile?x=1&y=2#frag")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if u.Scheme() != "file" {
		t.Fatalf("expected scheme file, got %q", u.Scheme())
	}
	if u.Path() != "/tmp/testfile" {
		t.Fatalf("expected path /tmp/testfile, got %q", u.Path())
	}
	if u.Host() != "" {
		t.Fatalf("expected empty host, got %q", u.Host())
	}
	if u.Fragment() != "frag" {
		t.Fatalf("expected fragment frag, got %q", u.Fragment())
	}
	if q := u.Query(); q.Get("x") != "1" || q.Get("y") != "2" {
		t.Fatalf("unexpected query map: %v", q)
	}

	str := u.String()
	parsed2, err := ParseURI(str)
	if err != nil {
		t.Fatalf("unexpected error parsing String() output: %v", err)
	}
	if parsed2.Scheme() != "file" || parsed2.Path() != "/tmp/testfile" {
		t.Fatalf("roundtrip ParseURI mismatch: %+v", parsed2)
	}
}

func TestURIPath_MarshalJSONAndText(t *testing.T) {
	u, xerr := ParseURI("file:///tmp/testfile?x=1")
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

func TestURIPath_BackendAs(t *testing.T) {
	stub := &testBackend{existsValue: true}
	u := &URIPath{backend: stub}

	asStub, err := BackendAs[*testBackend](u)
	if err != nil {
		t.Fatalf("expected no error from BackendAs[*testBackend], got %v", err)
	}
	if asStub != stub {
		t.Fatal("expected the same backend from BackendAs")
	}

	_, err = BackendAs[*FileBackend](u)
	if err == nil {
		t.Fatal("expected error from BackendAs[*FileBackend] when backend type mismatch")
	}
}

func TestURIPath_DelegatedMethods(t *testing.T) {
	stub := &testBackend{existsValue: true}
	u := &URIPath{backend: stub}

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

	if u.Backend() != stub {
		t.Fatal("Expected Backend() to return underlying backend")
	}
}

func TestParseURI_QueryAndFragment(t *testing.T) {
	raw := "file:///tmp/foo?val=5#bar"
	u, err := ParseURI(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Query()["val"][0] != "5" {
		t.Fatalf("expected query val=5, got %v", u.Query())
	}
	if u.Fragment() != "bar" {
		t.Fatalf("expected fragment bar, got %q", u.Fragment())
	}
}

func TestParseURI_RegisterBackend(t *testing.T) {
	// use a temporary scheme to avoid clashing with existing backends.
	RegisterBackend("testscheme", func(uri *URIPath, options ...BackendOption) (URIPathBackend, xerrors.Error) {
		return &testBackend{}, nil
	})

	u, err := ParseURI("testscheme://host/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Scheme() != "testscheme" {
		t.Fatalf("expected scheme testscheme, got %q", u.Scheme())
	}
}

func TestParseURI_SchemeLowercase(t *testing.T) {
	u, err := ParseURI("FiLe:///tmp/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Scheme() != "file" {
		t.Fatalf("expected scheme file (lowercased), got %q", u.Scheme())
	}
}

func TestGetQueryOptionValue_FromURLValues(t *testing.T) {
	query := url.Values{}
	query.Set("boolkey", "true")
	value := GetQueryOptionValue(false, "boolkey", query)
	if value != true {
		t.Fatalf("expected true, got %v", value)
	}
}

func TestParseURI_NoSchemeAbsolutePath(t *testing.T) {
	// absolute path without scheme should default to file backend
	u, err := ParseURI("/tmp/absolute/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if u.Scheme() != "file" {
		t.Fatalf("expected scheme file (default), got %q", u.Scheme())
	}
	if u.Path() != "/tmp/absolute/path" {
		t.Fatalf("expected path /tmp/absolute/path, got %q", u.Path())
	}
}

func TestParseURI_NoSchemeRelativePath(t *testing.T) {
	// relative path without scheme should default to file backend
	u, err := ParseURI("relative/path/file.txt", WithBackendOption("rel_root", "/root"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if u.Scheme() != "file" {
		t.Fatalf("expected scheme file (default), got %q", u.Scheme())
	}
	if u.Path() != "/root/relative/path/file.txt" {
		t.Fatalf("expected path /root/relative/path/file.txt, got %q", u.Path())
	}
}

func TestParseURI_WindowsPath(t *testing.T) {
	// Windows-style path (backslashes will be parsed as part of the path)
	u, err := ParseURI("file:///C:/Users/test/file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if u.Scheme() != "file" {
		t.Fatalf("expected scheme file, got %q", u.Scheme())
	}

	expectedPath := "/C:/Users/test/file.txt"
	if runtime.GOOS == "windows" {
		expectedPath = "C:\\Users\\test\\file.txt"
	}
	if u.Path() != expectedPath {
		t.Fatalf("expected path %s, got %q", expectedPath, u.Path())
	}
}

func TestParseURI_RelativeLinuxPathWithFileScheme(t *testing.T) {
	// relative path with explicit file scheme
	u, err := ParseURI("file://relative/path/to/file.txt", WithBackendOption("rel_root", "/root"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if u.Scheme() != "file" {
		t.Fatalf("expected scheme file, got %q", u.Scheme())
	}
	if u.Path() != "/root/relative/path/to/file.txt" {
		t.Fatalf("expected path /root/relative/path/to/file.txt, got %q", u.Path())
	}
	if u.Host() != "" {
		t.Fatalf("expected empty host, got %q", u.Host())
	}
}
func TestParseURI_RelativeLinuxPathWithNoWorkingDir(t *testing.T) {
	// save the original working directory so we can go back
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get initial wd: %v", err)
	}
	defer os.Chdir(originalWd)

	// create a temporary directory
	tmpDir, err := os.MkdirTemp("", "vanishing-dir")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// change into that directory
	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	// remove all permissions
	err = os.Chmod(tmpDir, 0000)
	if err != nil {
		t.Fatalf("failed to make directory inaccessible: %v", err)
	}

	// relative path without a root should try and use current directory and fail because there are no permissions
	_, err = ParseURI("file://relative/path/to/file.txt")
	if err == nil {
		t.Fatalf("expected an error but got none")
	}

	// try and cleanup
	os.RemoveAll(tmpDir)
}
