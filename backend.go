package uripath

import (
	"context"

	"go.innotegrity.dev/mod/xerrors"
)

// NewBackendFunc is a function that is used to create a new Backend instance for a given [URI].
//
// Options passed to this function will come from the [ParseURI] function and can be used in constructing the backend.
type NewBackendFunc func(ctx context.Context, uri *URI, options ...BackendOption) (Backend, xerrors.Error)

// Backend is the interface for a path URI backend.
//
// This interface is used by the [URI] struct to perform operations on the resource specified by the URI.
//
// If a custom backend is used, it must implement this interface. The easiest way to get started is by embedding the
// [BackendBase] struct into your custom backend and overriding the methods that you need to implement. [BackendBase]
// includes functionality to handle common backend options that callers can pass if they do not wish to specify them
// as part of the URI. Be sure to initialize the [BackendBase] struct before using it or its options.
type Backend interface {
	// Delete should delete the resource at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URI.Delete] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if the deletion failed.
	Delete(ctx context.Context, options ...BackendOption) xerrors.Error

	// Exists should check if the resource exists at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URI.Exists] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if there was an error checking for the existence of the resource.
	Exists(ctx context.Context, options ...BackendOption) (bool, xerrors.Error)

	// Get should retrieve the contents of the resource at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URI.Get] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if there was an error retrieving the contents of the resource.
	Get(ctx context.Context, options ...BackendOption) ([]byte, xerrors.Error)

	// List should list the resources at the given path.
	//
	// This function is typically used to list the contents of a folder. Use the recurse flag to determeine whether or
	// not to recurse into subfolders.
	//
	// The context and options passed to this function are the same as the ones passed to the [URI.List] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if there was an error listing the resources.
	List(ctx context.Context, recurse bool, options ...BackendOption) ([]string, xerrors.Error)

	// Options should return the common options that are stored with the backend.
	Options() map[string]any

	// Put should store the given data at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URI.Put] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if there was an error storing the data at the given path.
	Put(ctx context.Context, data []byte, options ...BackendOption) xerrors.Error

	// RemoveAllOptions shoud clear all common options stored with the backend.
	RemoveAllOptions()

	// RemoveOption should remove the option with the given key from the backend.
	//
	// The object itself is returned to allow for method chaining.
	RemoveOption(key string) Backend

	// ReplaceOptions should replace all common options stored with the backend with the given options.
	ReplaceOptions(options map[string]any)

	// SetOption should set the option with the given key to the given value.
	//
	// The object itself is returned to allow for method chaining.
	SetOption(key string, value any) Backend

	// URI should return the [URI] that the backend instance is associated with.
	URI() *URI
}

// BackendBase provides default implementations for Backend methods and can be embedded in specific backend
// implementations to reduce boilerplate code. It also provides common options handling functionality.
type BackendBase struct {
	// unexported variables
	options map[string]any // any common options for the file backend
	uri     *URI           // the URI object associated with this backend
}

// InitBackendBase initializes a [BackendBase] object with the provided options.
func InitBackendBase(uri *URI, options ...BackendOption) BackendBase {
	base := BackendBase{
		options: map[string]any{},
		uri:     uri,
	}
	for _, option := range options {
		option(&base)
	}

	return base
}

// Delete for this backend does nothing.
//
// Backends should override this method to provide actual delete functionality.
func (b *BackendBase) Delete(_ context.Context, _ ...BackendOption) xerrors.Error {
	return nil
}

// Exists for this backend does nothing and always returns false.
//
// Backends should override this method to provide actual existence checking functionality.
func (b *BackendBase) Exists(_ context.Context, _ ...BackendOption) (bool, xerrors.Error) {
	return false, nil
}

// Get for this backend does nothing and returns nil data.
//
// Backends should override this method to provide actual get functionality.
func (b *BackendBase) Get(_ context.Context, _ ...BackendOption) ([]byte, xerrors.Error) {
	return nil, nil
}

// List for this backend does nothing and returns a nil list.
//
// Backends should override this method to provide actual list functionality.
func (b *BackendBase) List(_ context.Context, _ bool, _ ...BackendOption) ([]string, xerrors.Error) {
	return nil, nil
}

// Put for this backend does nothing.
//
// Backends should override this method to provide actual put functionality.
func (b *BackendBase) Put(_ context.Context, _ []byte, _ ...BackendOption) xerrors.Error {
	return nil
}

// Options returns the backend options.
func (b *BackendBase) Options() map[string]any {
	return b.options
}

// RemoveAllOptions clears all options from the backend.
func (b *BackendBase) RemoveAllOptions() {
	b.options = map[string]any{}
}

// RemoveOption removes a specific option from the backend.
//
//nolint:ireturn // need to return interface to satisfy [Backend] interface
func (b *BackendBase) RemoveOption(key string) Backend {
	delete(b.options, key)

	return b
}

// ReplaceOptions replaces all options in the backend with the provided options.
func (b *BackendBase) ReplaceOptions(options map[string]any) {
	b.options = options
}

// SetOption sets a specific option in the backend.
//
//nolint:ireturn // need to return interface to satisfy [Backend] interface
func (b *BackendBase) SetOption(key string, value any) Backend {
	b.options[key] = value

	return b
}

// URI returns the URI object associated with this backend.
func (b *BackendBase) URI() *URI {
	return b.uri
}
