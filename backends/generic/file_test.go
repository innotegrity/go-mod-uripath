package generic_test

/*
func TestNewFileBackend(t *testing.T) {
	uri, xerr := uripath.ParseURI("file:///tmp/test.txt")
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	if backend == nil {
		t.Fatal("uripath.NewFileBackend returned nil")
	}

	// check that it's the correct type
	fileBackend, ok := backend.(*uripath.FileBackend)
	if !ok {
		t.Fatal("uripath.NewFileBackend did not return a *FileBackend")
	}

	if fileBackend.URIPath() != uri {
		t.Fatal("FileBackend URI not set correctly")
	}
}

func TestFileBackend_Delete(t *testing.T) {
	// create a temporary file
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")

	// create the file
	err := os.WriteFile(tempFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	uri, xerr := uripath.ParseURI("file://" + tempFile)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	deleteErr := backend.Delete(ctx)
	if deleteErr != nil {
		t.Fatalf("Failed to delete file: %v", deleteErr)
	}

	// check that the file no longer exists
	if _, err := os.Stat(tempFile); !os.IsNotExist(err) {
		t.Fatal("File still exists after deletion")
	}
}

func TestFileBackend_Delete_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "nonexistent.txt")

	uri, xerr := uripath.ParseURI("file://" + tempFile)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	deleteErr := backend.Delete(ctx)
	if deleteErr != nil {
		t.Fatalf("Failed to delete non-existent file: %v", deleteErr)
	}
}

func TestFileBackend_Delete_Directory(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "testdir")

	// create a directory with a file inside
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	tempFile := filepath.Join(testDir, "file.txt")
	err = os.WriteFile(tempFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file in dir: %v", err)
	}

	uri, xerr := uripath.ParseURI("file://" + testDir)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	deleteErr := backend.Delete(ctx)
	if deleteErr != nil {
		t.Fatalf("Failed to delete directory: %v", deleteErr)
	}

	// check that the directory no longer exists
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Fatal("Directory still exists after deletion")
	}
}

func TestFileBackend_Exists(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")

	// file doesn't exist yet
	uri, xerr := uripath.ParseURI("file://" + tempFile)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	exists, xerr := backend.Exists(ctx)
	if xerr != nil {
		t.Fatalf("Failed to check existence: %v", xerr)
	}
	if exists {
		t.Fatal("File should not exist")
	}

	// create the file
	err := os.WriteFile(tempFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	exists, xerr = backend.Exists(ctx)
	if xerr != nil {
		t.Fatalf("Failed to check existence: %v", xerr)
	}
	if !exists {
		t.Fatal("File should exist")
	}
}

func TestFileBackend_Get(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	content := []byte("Hello, World!")

	// create the file
	err := os.WriteFile(tempFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	uri, xerr := uripath.ParseURI("file://" + tempFile)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	data, xerr := backend.Get(ctx)
	if xerr != nil {
		t.Fatalf("Failed to get file content: %v", xerr)
	}

	if string(data) != string(content) {
		t.Fatalf("Content mismatch: got %q, want %q", string(data), string(content))
	}
}

func TestFileBackend_Get_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "nonexistent.txt")

	uri, xerr := uripath.ParseURI("file://" + tempFile)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	_, xerr = backend.Get(ctx)
	if xerr == nil {
		t.Fatal("Expected error when getting non-existent file")
	}

	// check error code is a expected
	if xerr.Code() != uripath.BackendGetError {
		t.Fatalf("Expected BackendGetError, got %d", xerr.Code())
	}
}

func TestFileBackend_Put(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	content := []byte("Hello, World!")

	uri, xerr := uripath.ParseURI("file://" + tempFile)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	xerr = backend.Put(ctx, content)
	if xerr != nil {
		t.Fatalf("Failed to put file content: %v", xerr)
	}

	// verify the file was created with correct content
	data, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	if string(data) != string(content) {
		t.Fatalf("Content mismatch: got %q, want %q", string(data), string(content))
	}
}

func TestFileBackend_Put_WithPermissions(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	content := []byte("Hello, World!")

	uri, xerr := uripath.ParseURI("file://" + tempFile)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	// set custom permissions
	ctx := context.Background()
	xerr = backend.Put(ctx, content, uripath.WithBackendOption("file_mode", os.FileMode(0600)),
		uripath.WithBackendOption("dir_mode", os.FileMode(0700)))
	if xerr != nil {
		t.Fatalf("Failed to put file content: %v", xerr)
	}

	// check file permissions
	info, err := os.Stat(tempFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	expectedMode := os.FileMode(0600)
	if info.Mode().Perm() != expectedMode {
		t.Fatalf("File permissions incorrect: got %v, want %v", info.Mode().Perm(), expectedMode)
	}
}

func TestFileBackend_List(t *testing.T) {
	tempDir := t.TempDir()

	// create some files and directories
	err := os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("content2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}
	err = os.MkdirAll(filepath.Join(tempDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "subdir", "file3.txt"), []byte("content3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file3: %v", err)
	}

	uri, xerr := uripath.ParseURI("file://" + tempDir)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()

	// test non-recursive list
	paths, xerr := backend.List(ctx, false)
	if xerr != nil {
		t.Fatalf("Failed to list directory: %v", xerr)
	}

	// should contain file1.txt, file2.txt, subdir
	expected := map[string]bool{
		filepath.Join(tempDir, "file1.txt"): true,
		filepath.Join(tempDir, "file2.txt"): true,
		filepath.Join(tempDir, "subdir"):    true,
	}

	if len(paths) != len(expected) {
		t.Fatalf("Expected %d paths, got %d: %v", len(expected), len(paths), paths)
	}

	for _, path := range paths {
		if !expected[path] {
			t.Fatalf("Unexpected path: %s", path)
		}
	}

	// test recursive list
	paths, xerr = backend.List(ctx, true)
	if xerr != nil {
		t.Fatalf("Failed to list directory recursively: %v", xerr)
	}

	expectedRecursive := map[string]bool{
		filepath.Join(tempDir, "file1.txt"):           true,
		filepath.Join(tempDir, "file2.txt"):           true,
		filepath.Join(tempDir, "subdir"):              true,
		filepath.Join(tempDir, "subdir", "file3.txt"): true,
	}

	if len(paths) != len(expectedRecursive) {
		t.Fatalf("Expected %d paths, got %d: %v", len(expectedRecursive), len(paths), paths)
	}

	for _, path := range paths {
		if !expectedRecursive[path] {
			t.Fatalf("Unexpected path: %s", path)
		}
	}
}

func TestFileBackend_List_NonExistentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "nonexistent")

	uri, xerr := uripath.ParseURI("file://" + nonExistentDir)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	_, xerr = backend.List(ctx, false)
	if xerr == nil {
		t.Fatal("Expected error when listing non-existent directory")
	}

	if xerr.Code() != uripath.BackendListError {
		t.Fatalf("Expected BackendListError, got %d", xerr.Code())
	}
}

func TestFileBackend_Delete_PathError(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "readonly")
	tempFile := filepath.Join(testDir, "test.txt")

	// create directory and file
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	err = os.WriteFile(tempFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// make directory read-only (no write permission)
	err = os.Chmod(testDir, 0555) // r-x for owner, no permissions for group/others
	if err != nil {
		t.Fatalf("Failed to make directory read-only: %v", err)
	}

	// defer cleanup - restore permissions so temp dir can be cleaned up
	defer os.Chmod(testDir, 0755)

	uri, xerr := uripath.ParseURI("file://" + tempFile)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	deleteErr := backend.Delete(ctx)
	if deleteErr == nil {
		t.Fatal("Expected error when deleting file in read-only directory")
	}

	// check that it's a BackendDeleteError
	if deleteErr.Code() != uripath.BackendDeleteError {
		t.Fatalf("Expected BackendDeleteError, got %d", deleteErr.Code())
	}
}

/*
func TestFileBackend_Exists_Error(t *testing.T) {
	tempDir := t.TempDir()
	testDir := filepath.Join(tempDir, "inaccessible")
	tempFile := filepath.Join(testDir, "test.txt")

	// create directory and file
	err := os.MkdirAll(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	err = os.WriteFile(tempFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// make directory inaccessible (no permissions)
	err = os.Chmod(testDir, 0000)
	if err != nil {
		t.Fatalf("Failed to make directory inaccessible: %v", err)
	}

	// defer cleanup - restore permissions so temp dir can be cleaned up
	defer os.Chmod(testDir, 0755)

	uri, xerr := uripath.ParseURI("file://" + tempFile)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	_, xerr = backend.Exists(ctx)
	if xerr == nil {
		t.Fatal("Expected error when checking existence of file in inaccessible directory")
	}

	// check that it's a BackendExistsError
	if xerr.Code() != uripath.BackendExistsError {
		t.Fatalf("Expected BackendExistsError, got %d", xerr.Code())
	}
}
*

func TestFileBackend_Put_MkdirAll_Error(t *testing.T) {
	tempDir := t.TempDir()
	parentDir := filepath.Join(tempDir, "readonly")
	targetDir := filepath.Join(parentDir, "nested")
	tempFile := filepath.Join(targetDir, "test.txt")
	content := []byte("Hello, World!")

	// create parent directory
	err := os.MkdirAll(parentDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create parent dir: %v", err)
	}

	// make parent directory read-only
	err = os.Chmod(parentDir, 0555)
	if err != nil {
		t.Fatalf("Failed to make parent directory read-only: %v", err)
	}

	// defer cleanup - restore permissions so temp dir can be cleaned up
	defer os.Chmod(parentDir, 0755)

	uri, xerr := uripath.ParseURI("file://" + tempFile)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	xerr = backend.Put(ctx, content)
	if xerr == nil {
		t.Fatal("Expected error when creating directory in read-only parent")
	}

	// check that it's a BackendPutError
	if xerr.Code() != uripath.BackendPutError {
		t.Fatalf("Expected BackendPutError, got %d", xerr.Code())
	}
}

func TestFileBackend_Put_WriteFile_Error(t *testing.T) {
	tempDir := t.TempDir()
	targetDir := filepath.Join(tempDir, "target")
	tempFile := filepath.Join(targetDir, "test.txt")
	content := []byte("Hello, World!")

	// create target directory
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create target dir: %v", err)
	}

	// make target directory read-only
	err = os.Chmod(targetDir, 0555)
	if err != nil {
		t.Fatalf("Failed to make target directory read-only: %v", err)
	}

	// defer cleanup - restore permissions so temp dir can be cleaned up
	defer os.Chmod(targetDir, 0755)

	uri, xerr := uripath.ParseURI("file://" + tempFile)
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewFileBackend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create FileBackend: %v", xerr)
	}

	ctx := context.Background()
	xerr = backend.Put(ctx, content)
	if xerr == nil {
		t.Fatal("Expected error when writing file to read-only directory")
	}

	// check that it's a BackendPutError
	if xerr.Code() != uripath.BackendPutError {
		t.Fatalf("Expected BackendPutError, got %d", xerr.Code())
	}
}

func TestFileBackend_Paths(t *testing.T) {
	nixPaths := map[string]string{
		"file:///root/path/to/file.txt": "/root/path/to/file.txt",
		"file://path/to/file.txt":       "/tmp/path/to/file.txt",
		"/root/path/to/file.txt":        "/root/path/to/file.txt",
		"path/to/file.txt":              "/tmp/path/to/file.txt",
	}
	winPaths := map[string]string{
		"file:///C:/foo/bar":          "C:\\foo\\bar",
		"file://C:/foo/bar":           "C:\\foo\\bar",
		"file:///C:foo/bar":           "C:\\temp\\foo\\bar",
		"file://C:foo/bar":            "C:\\temp\\foo\\bar",
		"/C:/foo/bar":                 "C:\\foo\\bar",
		"C:/foo/bar":                  "C:\\foo\\bar",
		"/C:foo/bar":                  "C:\\temp\\foo\\bar",
		"C:foo/bar":                   "C:\\temp\\foo\\bar",
		"C:\\foo\\bar":                "C:\\foo\\bar",
		"C:foo\\bar":                  "C:\\temp\\foo\\bar",
		"//server/share/foo/bar":      "\\\\server\\share\\foo\\bar",
		"\\\\server\\share\\foo\\bar": "\\\\server\\share\\foo\\bar",
	}

	for uri, expected := range nixPaths {
		u, err := uripath.ParseURI(uri, uripath.WithBackendOption("rel_root", "/tmp"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if u.Scheme() != "file" {
			t.Fatalf("expected scheme file, got %q", u.Scheme())
		}
		if u.Path() != expected {
			t.Fatalf("expected path %q, got %q", expected, u.Path())
		}
	}

	if runtime.GOOS == "windows" {
		for uri, _ := range winPaths {
			u, err := uripath.ParseURI(uri, uripath.WithBackendOption("rel_root", "C:\\temp"))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if u.Scheme() != "file" {
				t.Fatalf("expected scheme file, got %q", u.Scheme())
			}
			//if u.Path() != expected {
			//		t.Fatalf("expected path %q, got %q", expected, u.Path())
			//	}
		}
	}
}
*/
