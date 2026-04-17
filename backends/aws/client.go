package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"go.innotegrity.dev/mod/uripath"
	"go.innotegrity.dev/mod/xerrors"
)

// LoadDefaultConfig is a wrapper for [awsconfig.LoadDefaultConfig] that builds a "default" AWS client configuration
// using the given URI and options.
//
// The following options can be passed as query parameters **OR** via the list of backend options:
//   - api_access_key_id: the AWS access key ID to use for the client
//   - api_secret_access_key: the AWS secret access key to use for the client
//   - config_files: a comma-delimited list of files with AWS configuration settings
//   - config_profile: the name of the configuration profile to use (if other than "default")
//   - cred_files: a comma-delimited list of files with AWS credentials
//   - region: the AWS region to use for the client

// The following options can **only** be passed via the list of backend options:
//   - api_access_key_id: the AWS access key ID to use for the client
//   - api_secret_access_key: the AWS secret access key to use for the client
//
// Options passed via query parameters take precedence over those passed in via the list of backend options.
//
// If a username and password are both provided in the URI, the username is used as the access key ID and the
// password is used as the secret access key provided the optpion has not been passed via the list of backend options.
//
// This function may return an error with any of the following codes:
//   - [uripath.BackendInitError]: the AWS config could not be loaded
func LoadDefaultConfig(uri *uripath.URI, options ...uripath.BackendOption) (aws.Config, xerrors.Error) {
	var clientOptions []func(*awsconfig.LoadOptions) error

	// get query parameters and options from the URI
	queryParams := uri.Query
	accessKeyID := uripath.GetQueryOptionValue(uri.Username, "api_access_key_id", nil, options...)
	secretAccessKey := uripath.GetQueryOptionValue(uri.Password, "api_secret_access_key", nil, options...)
	configFilesStr := uripath.GetQueryOptionValue("", "config_files", queryParams, options...)
	configFiles := splitCommaList(configFilesStr)
	credFilesStr := uripath.GetQueryOptionValue("", "cred_files", queryParams, options...)
	credFiles := splitCommaList(credFilesStr)
	configProfile := uripath.GetQueryOptionValue("", "config_profile", queryParams, options...)
	region := uripath.GetQueryOptionValue("", "region", queryParams, options...)

	// configure AWS client options
	if accessKeyID != "" && secretAccessKey != "" {
		clientOptions = append(clientOptions, awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID, secretAccessKey, "")))
	}
	if len(configFiles) > 0 {
		clientOptions = append(clientOptions, awsconfig.WithSharedConfigFiles(configFiles))
	}
	if configProfile != "" {
		clientOptions = append(clientOptions, awsconfig.WithSharedConfigProfile(configProfile))
	}
	if len(credFiles) > 0 {
		clientOptions = append(clientOptions, awsconfig.WithSharedCredentialsFiles(credFiles))
	}
	if region != "" {
		clientOptions = append(clientOptions, awsconfig.WithRegion(region))
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(), clientOptions...)
	if err != nil {
		return aws.Config{}, xerrors.Wrapf(uripath.BackendInitError, err, "failed to load AWS config: %s", err.Error()).
			WithAttrs(map[string]any{
				"uri":               uri.String(),
				"access_key_id":     accessKeyID,
				"secret_access_key": secretAccessKey,
				"config_files":      configFiles,
				"config_profile":    configProfile,
				"cred_files":        credFiles,
				"region":            region,
			})
	}
	return cfg, nil
}

// splitCommaList splits a comma-delimited list of strings into a slice of strings eliminating any empty strings and
// trimming any whitespace.
func splitCommaList(list string) []string {
	result := []string{}
	for _, item := range strings.Split(list, ",") {
		item := strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}
