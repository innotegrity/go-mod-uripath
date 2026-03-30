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

// FileBackend implements [Backend] for local file system operations.
//
// The following rules apply to the URI for a [FileBackend]:
// - Absolute and relative paths without any scheme are supported for *nix and MacOS platforms.
// - Absolute and relative paths with the `file://` scheme are supported for *nix and MacOS platforms.
// - Windows paths **must** be prefixed with the `file://` scheme and should contain forward slashes. They will be
// translated automatically to backslashes.
// - For any other paths, either an error will be returned or files may not be correctly referenced.
type FileBackend struct {
	BackendBase
}

// NewFileBackend creates and initializes a new [FileBackend] object.
//
// If the URI path is not an absolute path, it is automatically converted to an absolute path when the backend is
// created.
//
// The following options are supported by this backend:
//   - "rel_root": the root path to prepend to any relative paths in the URI
//
// Duplicate options passed to a function will override any options set in the backend.
//
// This function will never return an error.
func NewFileBackend(uri *URIPath, options ...BackendOption) (Backend, xerrors.Error) {
	// convert the path to an absolute path
	if !filepath.IsAbs(uri.path) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, xerrors.Wrapf(BackendGetError, err, "failed to get current working directory: %s", err.Error())
		}
		uri.path = filepath.Clean(filepath.Join(GetFnOptionValue(cwd, "rel_root", nil, options...), uri.path))
	}

	return &FileBackend{
		BackendBase: InitBackendBase(uri, options...),
	}, nil
}

// Delete removes a file or directory (and its subdirectories) from the local filesystem.
//
// The context and options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendDeleteError]: the file or directory could not be deleted
func (f *FileBackend) Delete(ctx context.Context, options ...BackendOption) xerrors.Error {
	path := f.uri.Path()
	if err := os.RemoveAll(path); err != nil {
		return xerrors.Wrapf(BackendDeleteError, err, "failed to delete file '%s': %s", path, err.Error()).
			WithAttr("path", path)
	}
	return nil
}

// Exists checks if a file or directory exists on the local filesystem.
//
// The context and options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendExistsError]: the file or directory could not be checked
func (f *FileBackend) Exists(ctx context.Context, options ...BackendOption) (bool, xerrors.Error) {
	path := f.uri.Path()
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, xerrors.Wrapf(BackendExistsError, err, "failed to check for existence of '%s': %s", path,
			err.Error()).WithAttr("path", path)
	}
	return true, nil
}

// Get reads the content of a file from the local filesystem.
//
// The context and options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendGetError]: the file could not be read
func (f *FileBackend) Get(ctx context.Context, options ...BackendOption) ([]byte, xerrors.Error) {
	path := f.uri.Path()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, xerrors.Wrapf(BackendGetError, err, "failed to read file '%s': %s", path, err.Error()).
			WithAttr("path", path)
	}
	return data, nil
}

// List lists files and directories at the given path on the local filesystem.
//
// The context and options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendListError]: the directory could not be listed
func (f *FileBackend) List(ctx context.Context, recurse bool, options ...BackendOption) ([]string, xerrors.Error) {
	path := f.uri.Path()
	var paths []string
	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return xerrors.Wrapf(BackendListError, err, "failed to access path '%s': %s", walkPath, err.Error()).
				WithAttr("path", walkPath)
		}
		if walkPath == path || walkPath == "." || walkPath == ".." {
			return nil // skip the path itself and any parent listings
		}
		paths = append(paths, walkPath)
		if info.IsDir() && !recurse {
			return filepath.SkipDir // skip subdirectories if not recursing
		}
		return nil
	})
	if err != nil {
		return nil, xerrors.Wrapf(BackendListError, err, "failed to list directory '%s': %s", path, err.Error()).
			WithAttr("path", path)
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
func (f *FileBackend) Put(ctx context.Context, data []byte, options ...BackendOption) xerrors.Error {
	// parse options for file/directory permissions if provided, otherwise use the defaults
	dirPerm := GetFnOptionValue(os.FileMode(0755), "dir_mode", f.options, options...)
	filePerm := GetFnOptionValue(os.FileMode(0644), "file_mode", f.options, options...)

	// create parent directory if it doesn't exist
	path := f.uri.Path()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return xerrors.Wrapf(BackendPutError, err, "failed to create parent directory '%s': %s", dir, err.Error()).
			WithAttr("path", dir)
	}

	// write the file
	if err := os.WriteFile(path, data, filePerm); err != nil {
		return xerrors.Wrapf(BackendPutError, err, "failed to write file '%s': %s", path, err.Error()).
			WithAttr("path", path)
	}
	return nil
}
