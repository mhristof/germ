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
