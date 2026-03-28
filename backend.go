package uripath

import (
	"context"

	"go.innotegrity.dev/mod/xerrors"
)

// BackendBase provides default implementations for URIPathBackend methods and can be embedded in specific backend
// implementations to reduce boilerplate code. It also provides common options handling functionality.
type BackendBase struct {
	// unexported variables
	options map[string]any // any common options for the file backend
	uri     *URIPath       // the URI object associated with this backend
}

// Delete for this backend does nothing.
//
// Backends should override this method to provide actual delete functionality.
func (b *BackendBase) Delete(ctx context.Context, options ...map[string]any) xerrors.Error {
	return nil
}

// Exists for this backend does nothing and always returns false.
//
// Backends should override this method to provide actual existence checking functionality.
func (b *BackendBase) Exists(ctx context.Context, options ...map[string]any) (bool, xerrors.Error) {
	return false, nil
}

// Get for this backend does nothing and returns nil data.
//
// Backends should override this method to provide actual get functionality.
func (b *BackendBase) Get(ctx context.Context, options ...map[string]any) ([]byte, xerrors.Error) {
	return nil, nil
}

// List for this backend does nothing and returns a nil list.
//
// Backends should override this method to provide actual list functionality.
func (b *BackendBase) List(ctx context.Context, recurse bool, options ...map[string]any) ([]string, xerrors.Error) {
	return nil, nil
}

// Put for this backend does nothing.
//
// Backends should override this method to provide actual put functionality.
func (b *BackendBase) Put(ctx context.Context, data []byte, options ...map[string]any) xerrors.Error {
	return nil
}

// Options returns the backend options.
func (b *BackendBase) Options() map[string]any {
	return b.options
}

// RemoveAllOptions clears all options from the backend.
func (f *BackendBase) RemoveAllOptions() {
	f.options = map[string]any{}
}

// RemoveOption removes a specific option from the backend.
func (f *BackendBase) RemoveOption(key string) URIPathBackend {
	delete(f.options, key)
	return f
}

// ReplaceOptions replaces all options in the backend with the provided options.
func (f *BackendBase) ReplaceOptions(options map[string]any) {
	f.options = options
}

// SetOption sets a specific option in the backend.
func (f *BackendBase) SetOption(key string, value any) URIPathBackend {
	f.options[key] = value
	return f
}
