package generic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.innotegrity.dev/mod/xerrors"

	"go.innotegrity.dev/mod/uripath"
)

const (
	// FileScheme is the scheme for a file URI.
	FileScheme = "file"
)

func init() {
	_ = uripath.RegisterBackend(FileScheme, NewFileBackend)
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
// This function may return an error with any of the following codes:
//   - [uripath.BackendInitError]: the backend could not be initialized
func NewFileBackend(uri *uripath.URI, options ...uripath.BackendOption) (uripath.Backend, xerrors.Error) {
	// initialize the backend
	backend := &FileBackend{
		BackendBase: uripath.InitBackendBase(uri, options...),
	}
	fmt.Println("host:", uri.Host, "path:", uri.Path, "fragment:", uri.Fragment, "username:", uri.Username, "password:", uri.Password)
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
	return backend, nil
}

// Delete removes a file or directory (and its subdirectories) from the local filesystem.
//
// The context and options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendDeleteError]: the file or directory could not be deleted
func (f *FileBackend) Delete(ctx context.Context, options ...uripath.BackendOption) xerrors.Error {
	path := f.URI().Path
	if err := os.RemoveAll(path); err != nil {
		return xerrors.Wrapf(uripath.BackendDeleteError, err, "failed to delete file '%s': %s", path, err.Error()).
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
func (f *FileBackend) Exists(ctx context.Context, options ...uripath.BackendOption) (bool, xerrors.Error) {
	path := f.URI().Path
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, xerrors.Wrapf(uripath.BackendExistsError, err, "failed to check for existence of '%s': %s", path,
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
func (f *FileBackend) Get(ctx context.Context, options ...uripath.BackendOption) ([]byte, xerrors.Error) {
	path := f.URI().Path
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, xerrors.Wrapf(uripath.BackendGetError, err, "failed to read file '%s': %s", path, err.Error()).
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
func (f *FileBackend) List(ctx context.Context, recurse bool, options ...uripath.BackendOption) (
	[]string, xerrors.Error) {

	path := f.URI().Path
	var paths []string
	err := filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return xerrors.Wrapf(uripath.BackendListError, err, "failed to access path '%s': %s", walkPath, err.Error()).
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
		return nil, xerrors.Wrapf(uripath.BackendListError, err, "failed to list directory '%s': %s", path, err.Error()).
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
// The context passed to this function is not used.
//
// This function may return an error with any of the following codes:
//   - [BackendPutError]: the file could not be written
func (f *FileBackend) Put(ctx context.Context, data []byte, options ...uripath.BackendOption) xerrors.Error {
	// parse options for file/directory permissions if provided, otherwise use the defaults
	backendOptions := f.Options()
	dirPerm := uripath.GetFnOptionValue(os.FileMode(0755), "dir_mode", backendOptions, options...)
	filePerm := uripath.GetFnOptionValue(os.FileMode(0644), "file_mode", backendOptions, options...)

	// create parent directory if it doesn't exist
	path := f.URI().Path
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return xerrors.Wrapf(uripath.BackendPutError, err, "failed to create parent directory '%s': %s", dir,
			err.Error()).WithAttrs(map[string]any{
			"path":     dir,
			"dir_mode": dirPerm,
		})
	}

	// write the file
	if err := os.WriteFile(path, data, filePerm); err != nil {
		return xerrors.Wrapf(uripath.BackendPutError, err, "failed to write file '%s': %s", path, err.Error()).
			WithAttrs(map[string]any{
				"path":      path,
				"file_mode": filePerm,
			})
	}
	return nil
}
