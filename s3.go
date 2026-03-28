package uripath

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.innotegrity.dev/mod/xerrors"
)

const (
	// S3Scheme is the scheme for an Amazon S3 URI.
	S3Scheme = "s3"
)

func init() {
	RegisterBackend(S3Scheme, NewS3Backend)
}

// S3Backend implements [URIPathBackend] for Amazon S3 resources.
type S3Backend struct {
	BackendBase

	// unexported variables
	client *s3.Client
	bucket string
	key    string
}

// NewS3Backend creates an initializes a new [S3Backend] object.
//
// The following options can be passed as query parameters:
//   - region: the AWS region to use for the S3 client
//
// The following options can be passed in the options map:
//   - region: the AWS region to use for the S3 client
//
// Options passed in the query parameters take precedence over those in the options map.
//
// This function may return an error with any of the following codes:
//   - [InvalidParameter]: the URI is not valid
//   - [BackendInitError]: the S3 client could not be initialized
func NewS3Backend(uri *URIPath, options ...map[string]any) (URIPathBackend, xerrors.Error) {
	// setup required client variables
	bucket := uri.Host()
	if bucket == "" {
		return nil, xerrors.Newf(InvalidParameter, "s3 URI must include bucket name")
	}
	key := strings.TrimPrefix(uri.Path(), "/")

	// process options
	// TODO: support additional AWS client options
	queryParams := uri.Query()
	region := GetQueryOptionValue("", "region", queryParams, options...)

	// configure AWS client
	cfgOpts := []func(*awsconfig.LoadOptions) error{}
	if region != "" {
		cfgOpts = append(cfgOpts, awsconfig.WithRegion(region))
	}
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), cfgOpts...)
	if err != nil {
		return nil, xerrors.Wrapf(BackendGetError, err, "failed to load AWS config: %s", err.Error())
	}

	client := s3.NewFromConfig(cfg)
	return &S3Backend{
		BackendBase: BackendBase{
			options: map[string]any{},
			uri:     uri,
		},
		client: client,
		bucket: bucket,
		key:    key,
	}, nil
}

// Delete removes an object from the S3 bucket.
//
// The context passed to this function is passed to the S3 client.
//
// The options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendDeleteError]: the object could not be deleted
func (s *S3Backend) Delete(ctx context.Context, options ...map[string]any) xerrors.Error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
	}
	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return xerrors.Wrapf(BackendDeleteError, err, "failed to delete S3 object '%s/%s': %s", s.bucket,
			s.key, err.Error())
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
func (s *S3Backend) Exists(ctx context.Context, options ...map[string]any) (bool, xerrors.Error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
	}
	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		var ea *types.NotFound
		if errors.As(err, &ea) {
			return false, nil
		}

		// fallback for S3-style missing object errors.
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, xerrors.Wrapf(BackendExistsError, err, "failed to check for existence of S3 object '%s/%s': %s",
			s.bucket, s.key, err.Error())
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
func (s *S3Backend) Get(ctx context.Context, options ...map[string]any) ([]byte, xerrors.Error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
	}
	out, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, xerrors.Wrapf(BackendGetError, err, "failed to get S3 object '%s/%s': %s", s.bucket, s.key,
			err.Error())
	}
	defer out.Body.Close()
	buf := new(bytes.Buffer)
	if _, readErr := buf.ReadFrom(out.Body); readErr != nil {
		return nil, xerrors.Wrapf(BackendGetError, readErr, "failed reading body of S3 object '%s/%s': %s", s.bucket,
			s.key, readErr.Error())
	}
	return buf.Bytes(), nil
}

// List returns a list of objects in the S3 bucket.
//
// The context passed to this function is passed to the S3 client.
//
// The options passed to this function are not used.
//
// This function may return an error with any of the following codes:
//   - [BackendListError]: the objects could not be listed
func (s *S3Backend) List(ctx context.Context, recurse bool, options ...map[string]any) ([]string, xerrors.Error) {
	prefix := s.key
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}
	if !recurse {
		input.Delimiter = aws.String("/")
	}

	var results []string
	for {
		out, err := s.client.ListObjectsV2(ctx, input)
		if err != nil {
			return nil, xerrors.Wrapf(BackendListError, err, "failed to list S3 path '%s/%s': %s", s.bucket, s.key,
				err.Error())
		}
		for _, obj := range out.Contents {
			if obj.Key == nil {
				continue
			}
			results = append(results, fmt.Sprintf("%s/%s", s.bucket, *obj.Key))
		}
		if out.IsTruncated == nil || !*out.IsTruncated {
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
func (s *S3Backend) Put(ctx context.Context, data []byte, options ...map[string]any) xerrors.Error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
		Body:   bytes.NewReader(data),
	}
	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return xerrors.Wrapf(BackendPutError, err, "failed to write S3 object '%s/%s': %s", s.bucket, s.key, err.Error())
	}
	return nil
}
