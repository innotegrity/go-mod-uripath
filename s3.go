package uripath

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
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

	// Bucket is the name of the S3 bucket assocated with the backend.
	Bucket string

	// Client is the S3 client to use for API calls.
	Client S3ClientAPI

	// Key is the object key associated with the backend.
	Key string
}

// NewS3Backend creates an initializes a new [S3Backend] object.
//
// You can use the API access key ID and the API secret access key in place of username and password, respectively,
// in the URI if you wish to, however, this is not recommended as it may expose credentials in your config file.
// Instead, use environment variables, shared config files or instance roles where possible.
//
// The following options can be passed as query parameters:
//   - config_files: a comma-delimited list of files with AWS configuration settings
//   - config_profile: the name of the configuration profile to use (if other than "default")
//   - cred_files: a comma-delimited list of files with AWS credentials
//   - region: the AWS region to use for the S3 client
//
// The following options can be passed in the options map:
//   - api_access_key_id: the AWS access key ID to use for the S3 client
//   - api_secret_access_key: the AWS secret access key to use for the S3 client
//   - config_files: a comma-delimited list of files with AWS configuration settings
//   - config_profile: the name of the configuration profile to use (if other than "default")
//   - cred_files: a comma-delimited list of files with AWS credentials
//   - region: the AWS region to use for the S3 client
//   - s3_client: an existing S3 client object
//
// Options passed in the query parameters take precedence over those in the options map.
//
// Duplicate options passed to a function will override any options set in the backend.
//
// This function may return an error with any of the following codes:
//   - [InvalidParameter]: the URI is not valid
//   - [BackendInitError]: the S3 client could not be initialized
func NewS3Backend(uri *URIPath, options ...BackendOption) (URIPathBackend, xerrors.Error) {
	base := InitBackendBase(uri, options...)

	// setup required client variables
	bucket := uri.Host()
	if bucket == "" {
		return nil, xerrors.Newf(InvalidParameter, "s3 URI must include bucket name")
	}
	key := strings.TrimPrefix(uri.Path(), "/")

	// process any options specified
	queryParams := uri.Query()
	aws_access_key_id := GetQueryOptionValue("", "api_access_key_id", nil, options...)
	if aws_access_key_id == "" {
		aws_access_key_id = uri.username // will be empty if unspecified
	}
	aws_secret_access_key := GetQueryOptionValue("", "api_secret_access_key", nil, options...)
	if aws_secret_access_key == "" {
		aws_access_key_id = uri.password // will be empty if unspecified
	}
	configFiles := GetQueryOptionValue("", "config_files", queryParams, options...)
	configFileList := []string{}
	for _, file := range strings.Split(configFiles, ",") {
		file = strings.TrimSpace(file)
		if file != "" {
			configFileList = append(configFileList, file)
		}
	}
	configProfile := GetQueryOptionValue("", "config_profile", queryParams, options...)
	credFiles := GetQueryOptionValue("", "cred_files", queryParams, options...)
	credFileList := []string{}
	for _, file := range strings.Split(credFiles, ",") {
		file = strings.TrimSpace(file)
		if file != "" {
			credFileList = append(credFileList, file)
		}
	}
	region := GetQueryOptionValue("", "region", queryParams, options...)

	// configure the AWS client
	var client S3ClientAPI
	if v, ok := base.options["s3_client"]; ok {
		if cli, ok := v.(S3ClientAPI); ok {
			client = cli
		}
	}
	if client == nil {
		cfgOpts := []func(*awsconfig.LoadOptions) error{}
		if aws_access_key_id != "" && aws_secret_access_key != "" {
			cfgOpts = append(cfgOpts, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				aws_access_key_id, aws_secret_access_key, "")))
		}
		if len(configFileList) > 0 {
			cfgOpts = append(cfgOpts, awsconfig.WithSharedConfigFiles(configFileList))
		}
		if configProfile != "" {
			cfgOpts = append(cfgOpts, awsconfig.WithSharedConfigProfile(configProfile))
		}
		if len(credFileList) > 0 {
			cfgOpts = append(cfgOpts, awsconfig.WithSharedCredentialsFiles(credFileList))
		}
		if region != "" {
			cfgOpts = append(cfgOpts, awsconfig.WithRegion(region))
		}
		cfg, err := awsconfig.LoadDefaultConfig(context.Background(), cfgOpts...)
		if err != nil {
			return nil, xerrors.Wrapf(BackendGetError, err, "failed to load AWS config: %s", err.Error())
		}
		client = s3.NewFromConfig(cfg)
	}

	return &S3Backend{
		BackendBase: base,
		Bucket:      bucket,
		Client:      client,
		Key:         key,
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
func (s *S3Backend) Delete(ctx context.Context, options ...BackendOption) xerrors.Error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
	}
	_, err := s.Client.DeleteObject(ctx, input)
	if err != nil {
		return xerrors.Wrapf(BackendDeleteError, err, "failed to delete S3 object '%s/%s': %s", s.Bucket,
			s.Key, err.Error())
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
func (s *S3Backend) Exists(ctx context.Context, options ...BackendOption) (bool, xerrors.Error) {
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
		return false, xerrors.Wrapf(BackendExistsError, err, "failed to check for existence of S3 object '%s/%s': %s",
			s.Bucket, s.Key, err.Error())
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
func (s *S3Backend) Get(ctx context.Context, options ...BackendOption) ([]byte, xerrors.Error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
	}
	out, err := s.Client.GetObject(ctx, input)
	if err != nil {
		return nil, xerrors.Wrapf(BackendGetError, err, "failed to get S3 object '%s/%s': %s", s.Bucket, s.Key,
			err.Error())
	}
	defer out.Body.Close()
	buf := new(bytes.Buffer)
	if _, readErr := buf.ReadFrom(out.Body); readErr != nil {
		return nil, xerrors.Wrapf(BackendGetError, readErr, "failed reading body of S3 object '%s/%s': %s", s.Bucket,
			s.Key, readErr.Error())
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
func (s *S3Backend) List(ctx context.Context, recurse bool, options ...BackendOption) ([]string, xerrors.Error) {
	prefix := s.Key
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.Bucket),
		Prefix: aws.String(prefix),
	}
	if !recurse {
		input.Delimiter = aws.String("/")
	}

	var results []string
	for {
		out, err := s.Client.ListObjectsV2(ctx, input)
		if err != nil {
			return nil, xerrors.Wrapf(BackendListError, err, "failed to list S3 path '%s/%s': %s", s.Bucket, s.Key,
				err.Error())
		}
		for _, obj := range out.Contents {
			if obj.Key == nil {
				continue
			}
			results = append(results, fmt.Sprintf("%s/%s", s.Bucket, *obj.Key))
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
func (s *S3Backend) Put(ctx context.Context, data []byte, options ...BackendOption) xerrors.Error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(s.Key),
		Body:   bytes.NewReader(data),
	}
	_, err := s.Client.PutObject(ctx, input)
	if err != nil {
		return xerrors.Wrapf(BackendPutError, err, "failed to write S3 object '%s/%s': %s", s.Bucket, s.Key, err.Error())
	}
	return nil
}
