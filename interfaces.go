package uripath

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.innotegrity.dev/mod/xerrors"
)

// NewBackendFunc is a function that is used to create a new Backend instance for a given URIPath.
//
// Options passed to this function will come from the [ParseURI] function and can be used in constructing the backend.
type NewBackendFunc func(uri *URIPath, options ...BackendOption) (Backend, xerrors.Error)

// Backend is the interface for a path URI backend.
//
// This interface is used by the [URIPath] struct to perform operations on the resource specified by the URI.
//
// If a custom backend is used, it must implement this interface. The easiest way to get started is by embedding the
// [BackendBase] struct into your custom backend and overriding the methods that you need to implement. [BackendBase]
// includes functionality to handle common backend options that callers can pass if they do not wish to specify them
// as part of the URI. Be sure to initialize the [BackendBase] struct before using it or its options.
type Backend interface {
	// Delete should delete the resource at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URIPath.Delete] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if the deletion failed.
	Delete(ctx context.Context, options ...BackendOption) xerrors.Error

	// Exists should check if the resource exists at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URIPath.Exists] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if there was an error checking for the existence of the resource.
	Exists(ctx context.Context, options ...BackendOption) (bool, xerrors.Error)

	// Get should retrieve the contents of the resource at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URIPath.Get] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if there was an error retrieving the contents of the resource.
	Get(ctx context.Context, options ...BackendOption) ([]byte, xerrors.Error)

	// List should list the resources at the given path.
	//
	// This function is typically used to list the contents of a folder. Use the recurse flag to determeine whether or
	// not to recurse into subfolders.
	//
	// The context and options passed to this function are the same as the ones passed to the [URIPath.List] function.
	// They can be used by the function to perform operations on the resource.
	//
	// This function should return an error if there was an error listing the resources.
	List(ctx context.Context, recurse bool, options ...BackendOption) ([]string, xerrors.Error)

	// Options should return the common options that are stored with the backend.
	Options() map[string]any

	// Put should store the given data at the given path.
	//
	// The context and options passed to this function are the same as the ones passed to the [URIPath.Put] function.
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

	// URIPath should return the [URIPath] that the backend instance is associated with.
	URIPath() *URIPath
}

// S3ClientAPI is used for interacting with the S3 API but allows us to mock it for testing.
type S3ClientAPI interface {
	// DeleteObject deletes the object with the given key from the bucket.
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)

	// GetObject returns the object with the given key from the bucket.
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)

	// HeadObject returns metadata about the object with the given key from the bucket.
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)

	// ListObjectsV2 returns a list of objects in the bucket.
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)

	// PutObject writes data to the object with the given key in the bucket.
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}
