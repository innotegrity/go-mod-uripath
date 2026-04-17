package aws

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"go.innotegrity.dev/mod/uripath"
	"go.innotegrity.dev/mod/xerrors"
)

func init() {
	// Minimal backend so ParseURI can build a *uripath.URI for LoadDefaultConfig without AWS/S3 setup.
	_ = uripath.RegisterBackend("awsclienttest", newClientTestBackend, true)
}

type clientTestBackend struct{ uripath.BackendBase }

func newClientTestBackend(uri *uripath.URI, options ...uripath.BackendOption) (uripath.Backend, xerrors.Error) {
	return &clientTestBackend{BackendBase: uripath.InitBackendBase(uri, options...)}, nil
}

func TestSplitCommaList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty", in: "", want: nil},
		{name: "only commas and spaces", in: " , , ", want: nil},
		{name: "single", in: "a", want: []string{"a"}},
		{name: "trimmed", in: " a , b , c ", want: []string{"a", "b", "c"}},
		{name: "no extra empty", in: "x,,y", want: []string{"x", "y"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitCommaList(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("splitCommaList(%q) len = %d, want %d: %#v", tt.in, len(got), len(tt.want), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("splitCommaList(%q)[%d] = %q, want %q", tt.in, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLoadDefaultConfig_StaticCredentialsAndRegion(t *testing.T) {
	raw := "awsclienttest://AKIATESTACCESSKEY:secretaccesskey@bucket/key?region=us-east-1"
	u, xerr := uripath.ParseURI(raw)
	if xerr != nil {
		t.Fatalf("ParseURI: %v", xerr)
	}

	cfg, xerr := LoadDefaultConfig(u)
	if xerr != nil {
		t.Fatalf("LoadDefaultConfig: %v", xerr)
	}
	if cfg.Region != "us-east-1" {
		t.Fatalf("Region = %q, want us-east-1", cfg.Region)
	}
	creds, err := cfg.Credentials.Retrieve(t.Context())
	if err != nil {
		t.Fatalf("Credentials.Retrieve: %v", err)
	}
	if creds.AccessKeyID != "AKIATESTACCESSKEY" {
		t.Fatalf("AccessKeyID = %q", creds.AccessKeyID)
	}
	if creds.SecretAccessKey != "secretaccesskey" {
		t.Fatalf("SecretAccessKey mismatch")
	}
}

func TestLoadDefaultConfig_BackendOptionsCredentials(t *testing.T) {
	// Backend options used during ParseURI apply to the registered backend only; LoadDefaultConfig must receive
	// the same options (as S3Backend does when calling LoadDefaultConfig(uri, options...)).
	raw := "awsclienttest://bucket/key?region=us-west-2"
	u, xerr := uripath.ParseURI(raw)
	if xerr != nil {
		t.Fatalf("ParseURI: %v", xerr)
	}

	opts := []uripath.BackendOption{
		uripath.WithBackendOption("api_access_key_id", "OPTIONKEY"),
		uripath.WithBackendOption("api_secret_access_key", "OPTIONSECRET"),
	}
	cfg, xerr := LoadDefaultConfig(u, opts...)
	if xerr != nil {
		t.Fatalf("LoadDefaultConfig: %v", xerr)
	}
	if cfg.Region != "us-west-2" {
		t.Fatalf("Region = %q", cfg.Region)
	}
	creds, err := cfg.Credentials.Retrieve(t.Context())
	if err != nil {
		t.Fatalf("Credentials.Retrieve: %v", err)
	}
	if creds.AccessKeyID != "OPTIONKEY" || creds.SecretAccessKey != "OPTIONSECRET" {
		t.Fatalf("expected credentials from backend options")
	}
}

func TestLoadDefaultConfig_QueryOverridesBackendOptions(t *testing.T) {
	raw := "awsclienttest://bucket/key?region=us-west-2"
	u, xerr := uripath.ParseURI(raw, uripath.WithBackendOption("region", "us-east-1"))
	if xerr != nil {
		t.Fatalf("ParseURI: %v", xerr)
	}

	cfg, xerr := LoadDefaultConfig(u)
	if xerr != nil {
		t.Fatalf("LoadDefaultConfig: %v", xerr)
	}
	if cfg.Region != "us-west-2" {
		t.Fatalf("Region = %q, want query us-west-2 to win over backend option", cfg.Region)
	}
}

func TestLoadDefaultConfig_ConfigFilesFromQuery(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "config")
	if err := os.WriteFile(validPath, []byte("[default]\nregion = eu-west-1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	q := url.Values{}
	q.Set("config_files", validPath)
	raw := "awsclienttest://bucket/key?" + q.Encode()
	u, xerr := uripath.ParseURI(raw)
	if xerr != nil {
		t.Fatalf("ParseURI: %v", xerr)
	}

	cfg, xerr := LoadDefaultConfig(u)
	if xerr != nil {
		t.Fatalf("LoadDefaultConfig: %v", xerr)
	}
	if cfg.Region != "eu-west-1" {
		t.Fatalf("Region = %q, want eu-west-1 from shared config", cfg.Region)
	}
}

func TestLoadDefaultConfig_UsernamePasswordFallback(t *testing.T) {
	// No api_* in query; URI userinfo supplies keys (LoadDefaultConfig uses GetQueryOptionValue defaults).
	raw := "awsclienttest://MYKEY:MYSSECRET@bucket/key?region=ap-south-1"
	u, xerr := uripath.ParseURI(raw)
	if xerr != nil {
		t.Fatalf("ParseURI: %v", xerr)
	}

	cfg, xerr := LoadDefaultConfig(u)
	if xerr != nil {
		t.Fatalf("LoadDefaultConfig: %v", xerr)
	}
	creds, err := cfg.Credentials.Retrieve(t.Context())
	if err != nil {
		t.Fatalf("Credentials.Retrieve: %v", err)
	}
	if creds.AccessKeyID != "MYKEY" || creds.SecretAccessKey != "MYSSECRET" {
		t.Fatalf("expected credentials from URI userinfo")
	}
	if cfg.Region != "ap-south-1" {
		t.Fatalf("Region = %q", cfg.Region)
	}
}
