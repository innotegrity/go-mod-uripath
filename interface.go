package uripath

import (
	"context"

	"go.innotegrity.dev/mod/xerrors"
)

// NewURIPathBackendFunc is a function that is used to create a new URIPathBackend instance for a given URIPath.
//
// Options passed to this function will come from the [ParseURI] function and can be used in constructing the backend.
type NewURIPathBackendFunc func(uri *URIPath, options ...map[string]any) (URIPathBackend, xerrors.Error)

// URIPathBackend is the interface for a path URI backend.
//
// This interface is used by the [URIPath] struct to perform operations on the resource specified by the URI.
//
// If a custom backend is used, it must implement this interface. The easiest way to get started is by embedding the
// [BackendBase] struct into your custom backend and overriding the methods that you need to implement. [BackendBase]
// includes functionality to handle common backend options that callers can pass if they do not wish to specify them
// as part of the URI. Be sure to initialize the [BackendBase] struct before using it or its options.
type URIPathBackend interface {
	// Delete should delete the resource at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URIPath.Delete] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if the deletion failed.
	Delete(ctx context.Context, options ...map[string]any) xerrors.Error

	// Exists should check if the resource exists at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URIPath.Exists] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if there was an error checking for the existence of the resource.
	Exists(ctx context.Context, options ...map[string]any) (bool, xerrors.Error)

	// Get should retrieve the contents of the resource at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URIPath.Get] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if there was an error retrieving the contents of the resource.
	Get(ctx context.Context, options ...map[string]any) ([]byte, xerrors.Error)

	// List should list the resources at the given path.
	//
	// This function is typically used to list the contents of a folder. Use the recurse flag to determeine whether or
	// not to recurse into subfolders.
	//
	// The context and options passed to this function are the same as the ones passed to the [URIPath.List] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if there was an error listing the resources.
	List(ctx context.Context, recurse bool, options ...map[string]any) ([]string, xerrors.Error)

	// Options should return the common options that are stored with the backend.
	Options() map[string]any

	// Put should store the given data at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URIPath.Put] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if there was an error storing the data at the given path.
	Put(ctx context.Context, data []byte, options ...map[string]any) xerrors.Error

	// RemoveAllOptions shoud clear all common options stored with the backend.
	RemoveAllOptions()

	// RemoveOption should remove the option with the given key from the backend.
	//
	// The object itself is returned to allow for method chaining.
	RemoveOption(key string) URIPathBackend

	// ReplaceOptions should replace all common options stored with the backend with the given options.
	ReplaceOptions(options map[string]any)

	// SetOption should set the option with the given key to the given value.
	//
	// The object itself is returned to allow for method chaining.
	SetOption(key string, value any) URIPathBackend
}
