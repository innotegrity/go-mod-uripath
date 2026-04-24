package generic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	urierrors "go.innotegrity.dev/mod/uripath/errors"
	"go.innotegrity.dev/mod/xerrors"

	"go.innotegrity.dev/mod/uripath"
)

const (
	// FileScheme is the scheme for a file URI.
	FileScheme = "file"

	// defaultDirPerm reflects the default permissions to use when creating directories.
	defaultDirPerm = os.FileMode(0o755)

	// defaultFilePerm reflects the default permissions to use when creating files.
	defaultFilePerm = os.FileMode(0o644)
)

func init() { //nolint:gochecknoinits // this is a special case where we want to register the backend at startup
	_ = uripath.RegisterBackend(context.Background(), FileScheme, NewFileBackend)
}

// FileBackend implements [Backend] for local file system operations.
//
// The following rules apply to the URI for a [FileBackend]:
//   - All URIs must start with file:// to be valid.
//   - For all opreating systems, all paths should use forward slashes (/) instead backslashes (\).
//   - For all operating systems, relative paths should not include a leading slash (eg: file://relative/path/to/file).
//   - For MacOS and Linux, absolute paths must include a 3rd slash (eg: file:///absolute/path/to/file).
//   - For Windows, absolute paths must start with a drive letter and a colon (eg: file://C:/absolute/path) or for UNC
//     share paths, 2 additional slashes followed by the server name followed by a slash and the share name followed
//     by a slash and the path (eg: file:////server/share/absolute/path).
type FileBackend struct {
	uripath.BackendBase

	// Path is the path to the file or directory associated with the backend.
	Path string
}

// NewFileBackend creates and initializes a new [FileBackend] object.
//
// URIs must follow the rules for a [FileBackend].
//
// If the URI path is not an absolute path, it is automatically converted to an absolute path when the backend is
// created.
//
// The following options are can be passed as query parameters **OR** via the list of backend options:
//   - "dir_mode": the file permissions for any directories that may get created by the backend (default: 0755)
//   - "file_mode": the file permissions for any file that may get created by the backend (default: 0644)
//   - "rel_root": the root path to prepend to any relative paths in the URI (default: current working directory)
//
// Options passed via query parameters take precedence over those passed in via the list of backend options.
//
// This function may return any one of the following errors:
//
//   - [urierrors.BackendInitError]: the backend could not be initialized
//
//nolint:ireturn // need to return an interface for [uripath.NewBackendFunc] function signature.
func NewFileBackend(_ context.Context, uri *uripath.URI, options ...uripath.BackendOption) (
	uripath.Backend, xerrors.Error,
) {
	// initialize the backend
	backend := &FileBackend{
		BackendBase: uripath.InitBackendBase(uri, options...),
	}
	//nolint:forbidigo // need to print the URI settings for debugging purposes.
	fmt.Println(
		"host:",
		uri.Host,
		"path:",
		uri.Path,
		"fragment:",
		uri.Fragment,
		"username:",
		uri.Username,
		"password:",
		uri.Password,
	)

	return backend, nil
}

// Delete removes a file or directory (and its subdirectories) from the local filesystem.
//
// There are no options for this function.
//
// This function may return any one of the following errors:
//   - [BackendDeleteError]: the file or directory could not be deleted
func (f *FileBackend) Delete(ctx context.Context, _ ...uripath.BackendOption) xerrors.Error {
	path := f.URI().Path

	err := os.RemoveAll(path)
	if err != nil {
		return urierrors.NewBackendDeleteError(ctx, err, "failed to delete file '%s': %s", path, err.Error()).
			WithAttr("path", path)
	}

	return nil
}

// Exists checks if a file or directory exists on the local filesystem.
//
// There are no options for this function.
//
// This function may return any one of the following errors:
//   - [BackendExistsError]: the file or directory could not be checked
func (f *FileBackend) Exists(ctx context.Context, _ ...uripath.BackendOption) (bool, xerrors.Error) {
	path := f.URI().Path

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, urierrors.NewBackendExistsError(ctx, err, "failed to check for existence of '%s': %s", path,
			err.Error()).WithAttr("path", path)
	}

	return true, nil
}

// Get reads the content of a file from the local filesystem.
//
// There are no options for this function.
//
// This function may return any one of the following errors:
//   - [BackendGetError]: the file could not be read
func (f *FileBackend) Get(ctx context.Context, _ ...uripath.BackendOption) ([]byte, xerrors.Error) {
	path := f.URI().Path

	data, err := os.ReadFile(path) //nolint:gosec // should be able to read any file - caller must validate
	if err != nil {
		return nil, urierrors.NewBackendGetError(ctx, err, "failed to read file '%s': %s", path, err.Error()).
			WithAttr("path", path)
	}

	return data, nil
}

// List lists files and directories at the given path on the local filesystem.
//
// There are no options for this function.
//
// This function may return any one of the following errors:
//   - [BackendListError]: the directory could not be listed
func (f *FileBackend) List(ctx context.Context, recurse bool, _ ...uripath.BackendOption) (
	[]string, xerrors.Error,
) {
	path := f.URI().Path

	var paths []string

	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return urierrors.NewBackendListError(ctx, err, "failed to access path '%s': %s", walkPath, err.Error()).
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
		return nil, urierrors.NewBackendListError(ctx, err, "failed to list directory '%s': %s", path, err.Error()).
			WithAttr("path", path)
	}

	return paths, nil
}

// Put writes content to a file on the local filesystem.
//
// The following options can be passed as options and take precedence to any stored in the backend:
//   - "dir_mode": the file permissions for the parent directory (default: 0755)
//   - "file_mode": the file permissions (default: 0644)
//
// This function may return any one of the following errors:
//   - [BackendPutError]: the file could not be written
func (f *FileBackend) Put(ctx context.Context, data []byte, options ...uripath.BackendOption) xerrors.Error {
	// parse options for file/directory permissions if provided, otherwise use the defaults
	backendOptions := f.Options()
	dirPerm := uripath.GetFnOptionValue(defaultDirPerm, "dir_mode", backendOptions, options...)
	filePerm := uripath.GetFnOptionValue(defaultFilePerm, "file_mode", backendOptions, options...)

	// create parent directory if it doesn't exist
	path := f.URI().Path
	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, dirPerm)
	if err != nil {
		return urierrors.NewBackendPutError(ctx, err, "failed to create parent directory '%s': %s", dir,
			err.Error()).WithAttrs(map[string]any{
			"path":     dir,
			"dir_mode": dirPerm,
		})
	}

	// write the file
	err = os.WriteFile(path, data, filePerm)
	if err != nil {
		return urierrors.NewBackendPutError(ctx, err, "failed to write file '%s': %s", path, err.Error()).
			WithAttrs(map[string]any{
				"path":      path,
				"file_mode": filePerm,
			})
	}

	return nil
}

/*
	// if there's a hostname, that most likely means that the there was a relative file path, so we need to fix this
	if uri.host != "" {
		fmt.Println("  -- host is not empty! --", uri.host)
		uri.path = filepath.Join(uri.host + uri.path)
		uri.host = ""
	}

	// for Windows file:///C:/foo/bar results in /C:/foo/bar so we need to fix this
	if regexp.MustCompile(`^/[a-zA-Z]:`).MatchString(uri.path) {
		fmt.Println("  -- path is windows with a leading slash! --", uri.path)
		uri.path = uri.path[1:]
	}

	// convert the path to an absolute path
	if !filepath.IsAbs(uri.path) {
		fmt.Println("  -- path is not absolute! --", uri.path)
		cwd, err := os.Getwd()
		if err != nil {
			return nil, xerrors.Wrapf(BackendInitError, err, "failed to get current working directory: %s", err.Error())
		}
		uri.path = filepath.Clean(filepath.Join(GetFnOptionValue(cwd, "rel_root", nil, options...), uri.path))
	}

	// clean the path
	uri.path = filepath.Clean(uri.path)

	// force Windows style paths
	if GetFnOptionValue(false, "force_windows_path", nil, options...) == true {
		fmt.Println("  -- force_windows_path is true! --", uri.path)
		uri.path = strings.ReplaceAll(uri.path, "/", "\\")
	}

	// make sure the path is valid
	/*
		if _, err := os.Stat(uri.path); err != nil {
			if !os.IsNotExist(err) {
				return nil, xerrors.Wrapf(BackendInitError, err, "failed to stat path '%s': %s", uri.path, err.Error()).
					WithAttr("path", uri.path)
			}
		}
	*
	fmt.Println("final path:", uri.path)
*/
