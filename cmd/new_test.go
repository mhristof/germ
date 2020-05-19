package cmd

import (
	"io/ioutil"

	"os"
	"path/filepath"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/mhristof/gterm/log"
	"gotest.tools/assert"
)

func TestHandleFile(t *testing.T) {
	var cases = []struct {
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
	}

	for _, test := range cases {
		file, cleanup := tempFile(test.contents, test.fileName)
		defer cleanup()

		assert.Equal(t, test.exp, handleFile(file), test.name)
	}
}

func tempFile(contents, name string) (string, func()) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Panic("Cannot create temp dir")
	}

	tmpfn := filepath.Join(dir, name)
	if err := ioutil.WriteFile(tmpfn, []byte(contents), 0666); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Panic("Cannot write to file")
	}
	return tmpfn, func() {
		os.RemoveAll(dir)
	}
}
