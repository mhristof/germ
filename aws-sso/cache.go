package awssso

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sso"
	"github.com/mhristof/germ/iterm"
)

type Entries []*Entry

type Entry struct {
	AccessToken string `json:"accessToken"`
	ExpiresAt   string `json:"expiresAt"`
	Region      string `json:"region"`
	StartURL    string `json:"startUrl"`
}

func (e *Entries) Accounts() iterm.Profiles {
	var prof iterm.Profiles
	mySession := session.Must(session.NewSession())

	// Create a SSO client from just a session.
	svc := sso.New(mySession)

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	for _, entry := range *e {
		res, err := svc.ListAccounts(&sso.ListAccountsInput{
			AccessToken: aws.String(entry.AccessToken),
		})
		if err != nil {
			continue
		}

		startURL, err := url.Parse(entry.StartURL)
		if err != nil {
			panic(err)
		}

		hostname := startURL.Hostname()
		nameReplacer := strings.NewReplacer(" ", "-", ".", "-")

		for _, account := range res.AccountList {
			name := nameReplacer.Replace(
				fmt.Sprintf("%s-%s", hostname, strings.ToLower(*account.AccountName)),
			)

			var config = map[string]string{
				"Command": fmt.Sprintf("/usr/bin/env AWS_PROFILE=%s /usr/bin/login -fp %s", name, user.Username),
			}

			profile := iterm.NewProfile(name, config)
			prof.Add(*profile)
		}
	}

	fmt.Println(fmt.Sprintf("prof: %+v", prof))

	return prof
}

func Load(path string) (ret Entries, err error) {
	if path == "" {
		path, err = os.UserHomeDir()
		if err != nil {
			panic(err)
		}

		path = filepath.Join(path, ".aws/sso/cache")
	}

	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		var entry Entry

		fullPath := filepath.Join(path, file.Name())

		contents, err := ioutil.ReadFile(fullPath)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(contents, &entry)
		if err != nil {
			panic(err)
		}

		ret = append(ret, &entry)
	}

	return ret, nil
}
