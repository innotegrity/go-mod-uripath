package uripath

import (
	"context"
	"os"
	"path/filepath"

	"go.innotegrity.dev/mod/xerrors"
)

const (
	// FileScheme is the scheme for a file URI.
	FileScheme = "file"
)

func init() {
	RegisterBackend(FileScheme, NewFileBackend)
}

// FileBackend implements [URIPathBackend] for local file system operations.
type FileBackend struct {
	BackendBase
}

// NewFileBackend creates and initializes a new [FileBackend] object.
//
// The options passed to this function are not used.
//
// This function will never return an error.
func NewFileBackend(uri *URIPath, options ...map[string]any) (URIPathBackend, xerrors.Error) {
	return &FileBackend{
		BackendBase: BackendBase{
			options: map[string]any{},
			uri:     uri,
		},
	}, nil
}

// Delete removes a file or directory (and its subdirectories) from the local filesystem.
//
// The context and options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendDeleteError]: the file or directory could not be deleted
func (f *FileBackend) Delete(ctx context.Context, options ...map[string]any) xerrors.Error {
	if err := os.RemoveAll(f.uri.Path()); err != nil {
		return xerrors.Wrapf(BackendDeleteError, err, "failed to delete file '%s': %s", f.uri.Path(), err.Error()).
			WithAttr("path", f.uri.Path())
	}
	return nil
}

// Exists checks if a file or directory exists on the local filesystem.
//
// The context and options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendExistsError]: the file or directory could not be checked
func (f *FileBackend) Exists(ctx context.Context, options ...map[string]any) (bool, xerrors.Error) {
	_, err := os.Stat(f.uri.Path())
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, xerrors.Wrapf(BackendExistsError, err, "failed to check for existence of '%s': %s", f.uri.Path(),
			err.Error()).WithAttr("path", f.uri.Path())
	}
	return true, nil
}

// Get reads the content of a file from the local filesystem.
//
// The context and options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendGetError]: the file could not be read
func (f *FileBackend) Get(ctx context.Context, options ...map[string]any) ([]byte, xerrors.Error) {
	data, err := os.ReadFile(f.uri.Path())
	if err != nil {
		return nil, xerrors.Wrapf(BackendGetError, err, "failed to read file '%s': %s", f.uri.Path(), err.Error()).
			WithAttr("path", f.uri.Path())
	}
	return data, nil
}

// List lists files and directories at the given path on the local filesystem.
//
// The context and options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendListError]: the directory could not be listed
func (f *FileBackend) List(ctx context.Context, recurse bool, options ...map[string]any) ([]string, xerrors.Error) {
	var paths []string
	err := filepath.Walk(f.uri.Path(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return xerrors.Wrapf(BackendListError, err, "failed to access path '%s': %s", path, err.Error()).
				WithAttr("path", path)
		}
		if path == f.uri.Path() || path == "." || path == ".." {
			return nil // skip the path itself and any parent listings
		}
		paths = append(paths, path)
		if info.IsDir() && !recurse {
			return filepath.SkipDir // skip subdirectories if not recursing
		}
		return nil
	})
	if err != nil {
		return nil, xerrors.Wrapf(BackendListError, err, "failed to list directory '%s': %s", f.uri.Path(), err.Error()).
			WithAttr("path", f.uri.Path())
	}
	return paths, nil
}

// Put writes content to a file on the local filesystem.
//
// You can specify the file permissions for the parent directory and file using the "dir_mode" and "file_mode"
// options, respectively.  These options can be specified directly with this call or// in the options map passed to
// the backend.  If both exist, the options passed to this function take precedence.
//
// The context passed to this function is not used.
//
// This function may return an error with any of the following codes:
//   - [BackendPutError]: the file could not be written
func (f *FileBackend) Put(ctx context.Context, data []byte, options ...map[string]any) xerrors.Error {
	// parse options for file/directory permissions if provided, otherwise use the defaults
	dirPerm := GetFnOptionValue(os.FileMode(0755), "dir_mode", f.options, options...)
	filePerm := GetFnOptionValue(os.FileMode(0644), "file_mode", f.options, options...)

	// create parent directory if it doesn't exist
	dir := filepath.Dir(f.uri.Path())
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return xerrors.Wrapf(BackendPutError, err, "failed to create parent directory '%s': %s", dir, err.Error()).
			WithAttr("path", dir)
	}

	// write the file
	if err := os.WriteFile(f.uri.Path(), data, filePerm); err != nil {
		return xerrors.Wrapf(BackendPutError, err, "failed to write file '%s': %s", f.uri.Path(), err.Error()).
			WithAttr("path", f.uri.Path())
	}
	return nil
}
