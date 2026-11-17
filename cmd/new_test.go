package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"gotest.tools/assert"
)

func TestHandleFile(t *testing.T) {
	cases := []struct {
		name     string
		fileName string
		contents string
		exp      string
	}{
		{
			name:     "aws credentials.csv",
			fileName: "credentials.csv",
			contents: heredoc.Doc(`
				User name,Password,Access key ID,Secret access key,Console login link
				testUser,,AKIAqqqqqqqqqqqqqqqq,7FOtqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq,https://111111111111.signin.aws.amazon.com/console
			`),
			exp: "export AWS_ACCESS_KEY_ID=AKIAqqqqqqqqqqqqqqqq AWS_SECRET_ACCESS_KEY=7FOtqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
		},
		{
			name:     "aws accessKeys.csv",
			fileName: "accessKeys.csv",
			contents: heredoc.Doc(`
				Access key ID,Secret access key
				AKIAqqqqqqqqqqqqqqqq,1f1GsZqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq
			`),
			exp: "export AWS_ACCESS_KEY_ID=AKIAqqqqqqqqqqqqqqqq AWS_SECRET_ACCESS_KEY=1f1GsZqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
		},
		{
			name:     "aws root key file",
			fileName: "rootkey.csv",
			contents: heredoc.Doc(`
				AWSAccessKeyId=AKIAqqqqqqqqqqqqqqqq
				AWSSecretKey=1f1GsZqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq
			`),
			exp: "export AWS_ACCESS_KEY_ID=AKIAqqqqqqqqqqqqqqqq AWS_SECRET_ACCESS_KEY=1f1GsZqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq",
		},
	}

	for _, test := range cases {
		file, cleanup := tempFile(t, test.contents, test.fileName)
		defer cleanup()

		assert.Equal(t, test.exp, handleFile(file), test.name)
	}
}

func TestSanitizeSecret(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		exp     string
		wantErr bool
	}{
		{name: "plain token", in: "glsa_abcDEF123_XYZ9", exp: "glsa_abcDEF123_XYZ9"},
		{name: "trailing newline trimmed", in: "glsa_token9\n", exp: "glsa_token9"},
		{name: "surrounding whitespace trimmed", in: "  glsa_token9  ", exp: "glsa_token9"},
		{name: "empty rejected", in: "   ", wantErr: true},
		{name: "non-ascii accent rejected", in: "glsa_withñaccent", wantErr: true},
		{name: "embedded tab rejected", in: "glsa_tab\tend", wantErr: true},
		{name: "math symbol artifact rejected", in: "√glsa_token", wantErr: true},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := sanitizeSecret(test.in)
			if test.wantErr {
				assert.Assert(t, err != nil, "expected an error")
				return
			}
			assert.NilError(t, err)
			assert.Equal(t, test.exp, got)
		})
	}
}

func tempFile(t *testing.T, contents, name string) (string, func()) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatal(err)
	}

	tmpfn := filepath.Join(dir, name)
	if err := ioutil.WriteFile(tmpfn, []byte(contents), 0666); err != nil {
		t.Fatal(err)
	}
	return tmpfn, func() {
		os.RemoveAll(dir)
	}
}
