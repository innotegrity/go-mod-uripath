package aws

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.innotegrity.dev/mod/xerrors"

	"go.innotegrity.dev/mod/uripath"
)

const (
	// S3Scheme is the scheme for an Amazon S3 URI.
	S3Scheme = "s3"
)

func init() {
	_ = uripath.RegisterBackend(S3Scheme, NewS3Backend)
}

// S3ClientAPI is used for interacting with the S3 API but allows us to mock it for testing.
type S3ClientAPI interface {
	// DeleteObject deletes the object with the given key from the bucket.
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (
		*s3.DeleteObjectOutput, error)

	// GetObject returns the object with the given key from the bucket.
	GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (
		*s3.GetObjectOutput, error)

	// HeadObject returns metadata about the object with the given key from the bucket.
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (
		*s3.HeadObjectOutput, error)

	// ListObjectsV2 returns a list of objects in the bucket.
	ListObjectsV2(ctx context.Context, params *s3.ListObjectsV2Input, optFns ...func(*s3.Options)) (
		*s3.ListObjectsV2Output, error)

	// PutObject writes data to the object with the given key in the bucket.
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (
		*s3.PutObjectOutput, error)
}

// S3Backend implements [uripath.Backend] for Amazon S3 resources.
//
// The following rules apply to the URI for an [S3Backend]:
//   - All URIs must start with s3:// to be valid.
//   - The "host" portion of the URL should be the S3 bucket name.
//   - The "path" portion of the URL should be the S3 key.
//   - You may use an API access key ID for the "username" portion of the URL, if desired.
//   - You may use an API secret key for the "password" portion of the URL, if desired.
//   - Using API keys directly in the "username" and "password" portions of the URL is not recommended, however,
//     as it exposes credentials in plaintext.  Use EC2 instance roles, shared config files or environment
//     variables where possible.
type S3Backend struct {
	uripath.BackendBase

	// Bucket is the name of the S3 bucket assocated with the backend.
	Bucket string

	// Client is the S3 client to use for API calls.
	Client S3ClientAPI

	// Key is the object key associated with the backend.
	Key string
}

// NewS3Backend creates an initializes a new [S3Backend] for the given URI and options.
//
// URIs must follow the rules for an [S3Backend].
//
// The following options can be passed as query parameters **OR** via the list of backend options:
//   - config_files: a comma-delimited list of files with AWS configuration settings
//   - config_profile: the name of the configuration profile to use (if other than "default")
//   - cred_files: a comma-delimited list of files with AWS credentials
//   - region: the AWS region to use for the S3 client
//
// The following options can **only** be passed via the list of backend options:
//   - api_access_key_id: the AWS access key ID to use for the S3 client
//   - api_secret_access_key: the AWS secret access key to use for the S3 client
//   - s3_client: an existing S3 client object
//
// Options passed via query parameters take precedence over those passed in via the list of backend options.
//
// This function may return an error with any of the following codes:
//   - [uripath.InvalidParameter]: the URI is not valid
//   - [uripath.BackendInitError]: the S3 client could not be initialized
func NewS3Backend(uri *uripath.URI, options ...uripath.BackendOption) (uripath.Backend, xerrors.Error) {
	// validate the URI settings
	bucket := uri.Host
	if bucket == "" {
		return nil, xerrors.Newf(uripath.InvalidParameter, "s3 URI must include the bucket name as the host").
			WithAttr("uri", uri.String())
	}
	key := strings.TrimPrefix(uri.Path, "/")

	// initialize the backend
	backend := &S3Backend{
		BackendBase: uripath.InitBackendBase(uri, options...),
		Bucket:      bucket,
		Key:         key,
	}

	// initialize the S3 client
	backendOptions := backend.Options()
	if v, ok := backendOptions["s3_client"]; ok {
		if cli, ok := v.(S3ClientAPI); ok {
			backend.Client = cli
		}
	} else {
		cfg, xerr := LoadDefaultConfig(uri, options...)
		if xerr != nil {
			return nil, xerr
		}
		backend.Client = s3.NewFromConfig(cfg)
	}
	return backend, nil
}

// Delete removes an object from the S3 bucket.
//
// The context passed to this function is passed to the S3 client.
//
// The options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendDeleteError]: the object could not be deleted
func (s *S3Backend) Delete(ctx context.Context, options ...uripath.BackendOption) xerrors.Error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
	}
	_, err := s.Client.DeleteObject(ctx, input)
	if err != nil {
		return xerrors.Wrapf(uripath.BackendDeleteError, err, "failed to delete S3 object '%s/%s': %s", s.Bucket,
			s.Key, err.Error()).WithAttrs(map[string]any{
			"bucket": s.Bucket,
			"key":    s.Key,
		})
	}
	return nil
}

// Exists checks if an object exists in the S3 bucket.
//
// The context passed to this function is passed to the S3 client.
//
// The options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendExistsError]: the object could not be checked
func (s *S3Backend) Exists(ctx context.Context, options ...uripath.BackendOption) (bool, xerrors.Error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
	}
	_, err := s.Client.HeadObject(ctx, input)
	if err != nil {
		var ea *types.NotFound
		if errors.As(err, &ea) {
			return false, nil
		}

		// fallback for S3-style missing object errors.
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, xerrors.Wrapf(uripath.BackendExistsError, err,
			"failed to check for existence of S3 object '%s/%s': %s", s.Bucket, s.Key, err.Error()).
			WithAttrs(map[string]any{
				"bucket": s.Bucket,
				"key":    s.Key,
			})
	}
	return true, nil
}

// Get returns the contents of an object in the S3 bucket.
//
// The context passed to this function is passed to the S3 client.
//
// The options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendGetError]: the object could not be read
func (s *S3Backend) Get(ctx context.Context, options ...uripath.BackendOption) ([]byte, xerrors.Error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
	}
	out, err := s.Client.GetObject(ctx, input)
	if err != nil {
		return nil, xerrors.Wrapf(uripath.BackendGetError, err, "failed to get S3 object '%s/%s': %s", s.Bucket, s.Key,
			err.Error()).WithAttrs(map[string]any{
			"bucket": s.Bucket,
			"key":    s.Key,
		})
	}
	defer func() {
		_ = out.Body.Close()
	}()
	buf := new(bytes.Buffer)
	if _, readErr := buf.ReadFrom(out.Body); readErr != nil {
		return nil, xerrors.Wrapf(uripath.BackendGetError, readErr, "failed reading body of S3 object '%s/%s': %s",
			s.Bucket, s.Key, readErr.Error()).WithAttrs(map[string]any{
			"bucket": s.Bucket,
			"key":    s.Key,
		})
	}
	return buf.Bytes(), nil
}

// List returns a list of objects in the S3 bucket.
//
// The context passed to this function is passed to the S3 client.
//
// The options passed to this function are not used.
//
// The recurse flag is not used for S3 buckets as S3 objects are not organized into directories.
//
// This function may return an error with any of the following codes:
//   - [BackendListError]: the objects could not be listed
func (s *S3Backend) List(ctx context.Context, recurse bool, options ...uripath.BackendOption) (
	[]string, xerrors.Error) {
	prefix := normalizeS3ListPrefix(s.Key)
	input := newListObjectsV2Input(s.Bucket, prefix, recurse)

	var results []string
	for {
		out, err := s.Client.ListObjectsV2(ctx, input)
		if err != nil {
			return nil, xerrors.Wrapf(uripath.BackendListError, err, "failed to list S3 path '%s/%s': %s", s.Bucket,
				s.Key, err.Error()).WithAttrs(map[string]any{
				"bucket": s.Bucket,
				"key":    s.Key,
			})
		}
		results = appendS3ListObjectKeys(results, s.Bucket, out.Contents)
		if !listObjectsV2HasMore(out) {
			break
		}
		input.ContinuationToken = out.NextContinuationToken
	}
	return results, nil
}

// Put writes data to an object in the S3 bucket.
//
// The context passed to this function is passed to the S3 client.
//
// The options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendPutError]: the object could not be written
func (s *S3Backend) Put(ctx context.Context, data []byte, options ...uripath.BackendOption) xerrors.Error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
		Body:   bytes.NewReader(data),
	}
	_, err := s.Client.PutObject(ctx, input)
	if err != nil {
		return xerrors.Wrapf(uripath.BackendPutError, err, "failed to write S3 object '%s/%s': %s", s.Bucket, s.Key,
			err.Error()).WithAttrs(map[string]any{
			"bucket": s.Bucket,
			"key":    s.Key,
		})
	}
	return nil
}

// appendS3ListObjectKeys appends "bucket/key" strings for each object with a non-nil key.
func appendS3ListObjectKeys(results []string, bucket string, contents []types.Object) []string {
	for _, obj := range contents {
		if obj.Key == nil {
			continue
		}
		results = append(results, fmt.Sprintf("%s/%s", bucket, *obj.Key))
	}
	return results
}

// listObjectsV2HasMore reports whether ListObjectsV2 pagination should continue.
func listObjectsV2HasMore(out *s3.ListObjectsV2Output) bool {
	return out.IsTruncated != nil && *out.IsTruncated
}

// newListObjectsV2Input builds a [s3.ListObjectsV2Input] for the given bucket, prefix, and recursion mode.
func newListObjectsV2Input(bucket, prefix string, recurse bool) *s3.ListObjectsV2Input {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}
	if !recurse {
		input.Delimiter = aws.String("/")
	}
	return input
}

// normalizeS3ListPrefix ensures a non-empty key uses a trailing slash so S3 prefix listing treats it as a logical directory.
func normalizeS3ListPrefix(key string) string {
	if key == "" || strings.HasSuffix(key, "/") {
		return key
	}
	return key + "/"
}
