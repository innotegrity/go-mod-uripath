package aws_test

/*
func TestNewS3Backend(t *testing.T) {
	// test valid s3 uri
	uri, xerr := uripath.ParseURI("s3://mybucket/path/to/file.txt")
	if xerr != nil {
		t.Fatalf("Failed to parse valid S3 URI: %v", xerr)
	}

	backend, xerr := uripath.NewS3Backend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create S3Backend: %v", xerr)
	}

	if backend == nil {
		t.Fatal("uripath.NewS3Backend returned nil")
	}

	// check that it's the correct type
	s3Backend, ok := backend.(*uripath.S3Backend)
	if !ok {
		t.Fatal("uripath.NewS3Backend did not return a *uripath.S3Backend")
	}

	if s3Backend.Bucket != "mybucket" {
		t.Fatalf("Expected Bucket 'mybucket', got '%s'", s3Backend.Bucket)
	}

	if s3Backend.Key != "path/to/file.txt" {
		t.Fatalf("Expected key 'path/to/file.txt', got '%s'", s3Backend.Key)
	}
}

func TestNewS3Backend_InvalidURI_NoBucket(t *testing.T) {
	// test s3 uri without Bucket
	_, xerr := uripath.ParseURI("s3:///path/to/file.txt")
	if xerr == nil {
		t.Fatal("Expected error for S3 URI without Bucket")
	}

	if xerr.Code() != uripath.InvalidParameter {
		t.Fatalf("Expected InvalidParameter error, got %d", xerr.Code())
	}
}

func TestNewS3Backend_WithRegion(t *testing.T) {
	// test s3 uri with region in query
	uri, xerr := uripath.ParseURI("s3://mybucket/path/to/file.txt?region=us-west-2")
	if xerr != nil {
		t.Fatalf("Failed to parse S3 URI with region: %v", xerr)
	}

	backend, xerr := uripath.NewS3Backend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create S3Backend with region: %v", xerr)
	}

	if backend == nil {
		t.Fatal("uripath.NewS3Backend with region returned nil")
	}

	s3Backend, ok := backend.(*uripath.S3Backend)
	if !ok {
		t.Fatal("uripath.NewS3Backend did not return a *uripath.S3Backend")
	}

	if s3Backend.Bucket != "mybucket" {
		t.Fatalf("Expected Bucket 'mybucket', got '%s'", s3Backend.Bucket)
	}
}

func TestNewS3Backend_RootKey(t *testing.T) {
	// test s3 uri with root key (empty path)
	uri, xerr := uripath.ParseURI("s3://mybucket")
	if xerr != nil {
		t.Fatalf("Failed to parse S3 URI with root key: %v", xerr)
	}

	backend, xerr := uripath.NewS3Backend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create S3Backend with root key: %v", xerr)
	}

	s3Backend, ok := backend.(*uripath.S3Backend)
	if !ok {
		t.Fatal("uripath.NewS3Backend did not return a *uripath.S3Backend")
	}

	if s3Backend.Key != "" {
		t.Fatalf("Expected empty key for root path, got '%s'", s3Backend.Key)
	}
}

func TestNewS3Backend_KeyWithLeadingSlash(t *testing.T) {
	// test s3 uri with key that has leading slash
	uri, xerr := uripath.ParseURI("s3://mybucket//path/to/file.txt")
	if xerr != nil {
		t.Fatalf("Failed to parse S3 URI with leading slash: %v", xerr)
	}

	backend, xerr := uripath.NewS3Backend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create S3Backend with leading slash: %v", xerr)
	}

	s3Backend, ok := backend.(*uripath.S3Backend)
	if !ok {
		t.Fatal("uripath.NewS3Backend did not return a *uripath.S3Backend")
	}

	// the leading slash should be trimmed
	if s3Backend.Key != "/path/to/file.txt" {
		t.Fatalf("Expected key '/path/to/file.txt', got '%s'", s3Backend.Key)
	}
}

func TestS3Backend_Delete_WithMockClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockS3ClientAPI(ctrl)
	mockClient.EXPECT().DeleteObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(&s3.DeleteObjectOutput{}, nil)

	uri, xerr := uripath.ParseURI("s3://mybucket/path/to/file.txt", uripath.WithBackendOption("s3_client", mockClient))
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	xerr = uri.Delete(context.Background(), nil)
	if xerr != nil {
		t.Fatalf("Expected nil error from Delete with mock client, got: %v", xerr)
	}
}

func TestS3Backend_Exists_WithMockClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockS3ClientAPI(ctrl)
	mockClient.EXPECT().HeadObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(&s3.HeadObjectOutput{}, nil)

	uri, xerr := uripath.ParseURI("s3://mybucket/path/to/file.txt", uripath.WithBackendOption("s3_client", mockClient))
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	exists, xerr := uri.Exists(context.Background(), nil)
	if xerr != nil {
		t.Fatalf("Expected nil error from Exists with mock client, got: %v", xerr)
	}
	if !exists {
		t.Fatal("Expected exists=true from mock client")
	}
}

func TestS3Backend_Exists_NotFoundWithMockClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockS3ClientAPI(ctrl)
	mockClient.EXPECT().HeadObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, &types.NotFound{})

	uri, xerr := uripath.ParseURI("s3://mybucket/path/to/file.txt", uripath.WithBackendOption("s3_client", mockClient))
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	exists, xerr := uri.Exists(context.Background(), nil)
	if xerr != nil {
		t.Fatalf("Expected nil error from Exists for not found, got: %v", xerr)
	}
	if exists {
		t.Fatal("Expected exists=false from mocked NotFound")
	}
}

func TestS3Backend_Get_WithMockClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockS3ClientAPI(ctrl)
	body := io.NopCloser(bytes.NewBufferString("hello"))
	mockClient.EXPECT().GetObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(&s3.GetObjectOutput{Body: body}, nil)

	uri, xerr := uripath.ParseURI("s3://mybucket/path/to/file.txt", uripath.WithBackendOption("s3_client", mockClient))
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	data, xerr := uri.Get(context.Background(), nil)
	if xerr != nil {
		t.Fatalf("Expected nil error from Get with mock client, got: %v", xerr)
	}
	if string(data) != "hello" {
		t.Fatalf("Expected content 'hello', got '%s'", string(data))
	}
}

func TestS3Backend_List_WithMockClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockS3ClientAPI(ctrl)
	mockClient.EXPECT().ListObjectsV2(gomock.Any(), gomock.Any(), gomock.Any()).Return(&s3.ListObjectsV2Output{
		Contents:              []types.Object{{Key: aws.String("path/to/file.txt")}},
		IsTruncated:           aws.Bool(false),
		NextContinuationToken: nil,
	}, nil)

	uri, xerr := uripath.ParseURI("s3://mybucket/path/to", uripath.WithBackendOption("s3_client", mockClient))
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	paths, xerr := uri.List(context.Background(), true, nil)
	if xerr != nil {
		t.Fatalf("Expected nil error from List with mock client, got: %v", xerr)
	}
	if len(paths) != 1 || paths[0] != "mybucket/path/to/file.txt" {
		t.Fatalf("Expected one path mybucket/path/to/file.txt, got %v", paths)
	}
}

func TestS3Backend_Put_WithMockClient(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockS3ClientAPI(ctrl)
	mockClient.EXPECT().PutObject(gomock.Any(), gomock.Any(), gomock.Any()).Return(&s3.PutObjectOutput{}, nil)

	uri, xerr := uripath.ParseURI("s3://mybucket/path/to/file.txt", uripath.WithBackendOption("s3_client", mockClient))
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	xerr = uri.Put(context.Background(), []byte("mock content"), nil)
	if xerr != nil {
		t.Fatalf("Expected nil error from Put with mock client, got: %v", xerr)
	}
}

func TestS3Backend_Get(t *testing.T) {
	// this test would require a mocked s3 client or actual aws s3 access
	t.Skip("Skipping S3 integration test - requires AWS credentials or mocked client")

	uri, xerr := uripath.ParseURI("s3://test-Bucket/test-key")
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewS3Backend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create S3Backend: %v", xerr)
	}

	ctx := context.Background()
	// this would get object content from s3
	data, xerr := backend.Get(ctx)
	// in a real test, you'd verify the content
	_ = data // placeholder
	_ = xerr // placeholder
}

func TestS3Backend_List(t *testing.T) {
	// this test would require a mocked s3 client or actual aws s3 access
	t.Skip("Skipping S3 integration test - requires AWS credentials or mocked client")

	uri, xerr := uripath.ParseURI("s3://test-Bucket/test-prefix/")
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewS3Backend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create S3Backend: %v", xerr)
	}

	ctx := context.Background()
	// this would list objects from s3
	paths, xerr := backend.List(ctx, false)
	// in a real test, you'd verify the results
	_ = paths // placeholder
	_ = xerr  // placeholder
}

func TestS3Backend_List_Recursive(t *testing.T) {
	// this test would require a mocked s3 client or actual aws s3 access
	t.Skip("Skipping S3 integration test - requires AWS credentials or mocked client")

	uri, xerr := uripath.ParseURI("s3://test-Bucket/test-prefix/")
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewS3Backend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create S3Backend: %v", xerr)
	}

	ctx := context.Background()
	// this would recursively list objects from s3
	paths, xerr := backend.List(ctx, true)
	// in a real test, you'd verify the results
	_ = paths // placeholder
	_ = xerr  // placeholder
}

func TestS3Backend_Put(t *testing.T) {
	// this test would require a mocked s3 client or actual aws s3 access
	t.Skip("Skipping S3 integration test - requires AWS credentials or mocked client")

	uri, xerr := uripath.ParseURI("s3://test-Bucket/test-key")
	if xerr != nil {
		t.Fatalf("Failed to parse URI: %v", xerr)
	}

	backend, xerr := uripath.NewS3Backend(uri)
	if xerr != nil {
		t.Fatalf("Failed to create S3Backend: %v", xerr)
	}

	ctx := context.Background()
	content := []byte("test content")
	// this would put object content to s3
	xerr = backend.Put(ctx, content)
	// in a real test, you'd check for success
	_ = xerr // placeholder
}
*/
