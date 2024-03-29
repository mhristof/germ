package iterm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/assert"
)

func TestTags(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]string
		result []string
	}{
		{
			name: "Source profile",
			config: map[string]string{
				"source_profile": "bar",
				"timestamps":     "false",
			},
			result: []string{
				"source-profile=bar",
				"bar",
			},
		},
		{
			name: "account from azure app id",
			config: map[string]string{
				"azure_app_id_uri": `https://signin.aws.amazon.com/saml\#123456789012`,
				"timestamps":       "false",
			},
			result: []string{
				"123456789012",
			},
		},
		{
			name: "azure_default_role_arn",
			config: map[string]string{
				"azure_default_role_arn": `arn:aws:iam::123456789012:role/thisRole`,
				"timestamps":             "false",
			},
			result: []string{
				"role/thisRole",
			},
		},
		{
			name: "comma separated tags",
			config: map[string]string{
				"Tags":       "this,that",
				"timestamps": "false",
			},
			result: []string{
				"this",
				"that",
			},
		},
		{
			name: "section with role_arn",
			config: map[string]string{
				"role_arn":   "arn:aws:iam::123456789012:role/name",
				"timestamps": "false",
			},
			result: []string{
				"123456789012",
			},
		},
		{
			name: "section with sso_account_id",
			config: map[string]string{
				"sso_account_id": "123456789012",
				"timestamps":     "false",
			},
			result: []string{
				"account=123456789012",
			},
		},
		{
			name: "section with timestamps",
			config: map[string]string{
				"timestamps": "true",
			},
			result: []string{
				time.Now().Format(time.RFC3339),
			},
		},
		{
			name: "section without timestamps",
			config: map[string]string{
				"timestamps": "false",
			},
			result: []string{},
		},
	}

	for _, test := range cases {
		fmt.Println(fmt.Sprintf("test: %+v", test))
		assert.Equal(t, test.result, Tags(test.config), test.name)
	}
}

func TestNewProfile(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]string
		eval   func(*Profile) bool
	}{
		{
			name:   "check resycle values",
			config: map[string]string{},
		},
		{
			name: "profile with cmd",
			config: map[string]string{
				"Command": "/bin/bash",
			},
			eval: func(p *Profile) bool {
				return (p.Command == "/bin/bash" && p.CustomCommand == "Yes")
			},
		},
		{
			name:   "profile without cmd",
			config: map[string]string{},
			eval: func(p *Profile) bool {
				return (p.Command == "" && p.CustomCommand == "")
			},
		},
		{
			name: "profile with comma separated tags in config",
			config: map[string]string{
				"Tags": "this,that",
			},
			eval: func(p *Profile) bool {
				return p.Tags[1] == "this" && p.Tags[2] == "that"
			},
		},
		{
			name:   "production profile (red)",
			config: map[string]string{},
			eval: func(p *Profile) bool {
				return p.BackgroundColor.RedComponent != 0
			},
		},
		{
			name:   "nonproduction profile (nonred)",
			config: map[string]string{},
			eval: func(p *Profile) bool {
				return p.BackgroundColor.RedComponent == 0
			},
		},
		{
			name: "Kubernetes profile",
			config: map[string]string{
				"Tags": "k8s",
			},
			eval: func(p *Profile) bool {
				return p.BackgroundColor.BlueComponent != 0
			},
		},
		{
			name: "custom badge text",
			config: map[string]string{
				"BadgeText": "awesome",
			},
			eval: func(p *Profile) bool {
				return p.BadgeText == "awesome"
			},
		},
		{
			name: "custom title setting",
			config: map[string]string{
				"AllowTitleSetting": "true",
			},
			eval: func(p *Profile) bool {
				return p.AllowTitleSetting == true
			},
		},
	}

	for _, test := range cases {
		profile := NewProfile(test.name, test.config)
		assert.Contains(t, []string{"Recycle"}, profile.CustomDirectory)
		if test.eval != nil {
			assert.True(t, test.eval(profile), test.name)
		}
	}
}

func TestUpdateKeyboardMaps(t *testing.T) {
	cases := []struct {
		name     string
		profiles Profiles
	}{
		{
			name: "aws profile, k8s with that profile as source",
			profiles: Profiles{
				Profiles: []Profile{
					{
						GUID: "awesomeAWSProfile",
						KeyboardMap: map[string]KeyboardMap{
							KeyboardSortcutAltA: {
								Text: "tada!",
							},
						},
					},
					{
						GUID: "k8s",
						Tags: []string{
							"k8s",
							"aws-profile=awesomeAWSProfile",
						},
						KeyboardMap: map[string]KeyboardMap{},
					},
				},
			},
		},
	}

	for _, test := range cases {
		test.profiles.UpdateKeyboardMaps()
		assert.Equal(t, test.profiles.Profiles[1].KeyboardMap[KeyboardSortcutAltA].Text, "tada!")
	}
}

func TestUpdateAWSSmartSelectionRules(t *testing.T) {
	cases := []struct {
		name     string
		profiles Profiles
		exp      []SmartSelectionRule
	}{
		{
			name: "One profile with an account tag",
			profiles: Profiles{
				Profiles: []Profile{
					{
						Name: "account 1",
						Tags: []string{"account=account1"},
					},
					{
						Name: "account 2",
					},
				},
			},
			exp: []SmartSelectionRule{
				{
					Actions: []SmartSelectionRuleAction{
						{
							Action:    2,
							Parameter: "osascript -e 'display notification \"account 1\" with title \"account1\"'",
							Title:     "Notify the AWS account name",
						},
					},
					Notes:     "AWS account ID for account 1",
					Precision: "very_high",
					Regex:     "account1",
				},
			},
		},
		{
			name: "profile with login- prefix",
			profiles: Profiles{
				Profiles: []Profile{
					{
						Name: "account 1",
						Tags: []string{"account=account1"},
					},
					{
						Name: "login-account 1",
						Tags: []string{"account=account1"},
					},
					{
						Name: "account 2",
					},
				},
			},
			exp: []SmartSelectionRule{
				{
					Actions: []SmartSelectionRuleAction{
						{
							Action:    2,
							Parameter: "osascript -e 'display notification \"account 1\" with title \"account1\"'",
							Title:     "Notify the AWS account name",
						},
					},
					Notes:     "AWS account ID for account 1",
					Precision: "very_high",
					Regex:     "account1",
				},
			},
		},
	}

	for _, test := range cases {
		test.profiles.UpdateAWSSmartSelectionRules()
		assert.Equal(t, test.exp, test.profiles.Profiles[1].SmartSelectionRules, test.name)
	}
}

func TestColors(t *testing.T) {
	cases := []struct {
		name    string
		profile Profile
		exp     func(Profile) bool
	}{
		{
			name: "Production profile",
			profile: Profile{
				Name: "prod",
			},
			exp: func(p Profile) bool {
				return p.BackgroundColor.RedComponent != 0
			},
		},
		{
			name: "Non Production profile",
			profile: Profile{
				Name: "nonprod",
			},
			exp: func(p Profile) bool {
				return p.BackgroundColor.RedComponent == 0
			},
		},
		{
			name: "Production profile",
			profile: Profile{
				Name: "prd",
			},
			exp: func(p Profile) bool {
				return p.BackgroundColor.RedComponent != 0
			},
		},
		{
			name: "Non Production profile",
			profile: Profile{
				Name: "nonprd",
			},
			exp: func(p Profile) bool {
				return p.BackgroundColor.RedComponent == 0
			},
		},
		{
			name: "Kubernets profile",
			profile: Profile{
				Name: "kubernetez",
				Tags: []string{
					"k8s",
				},
			},
			exp: func(p Profile) bool {
				return p.BackgroundColor.BlueComponent != 0
			},
		},
		{
			name: "Generic profile",
			profile: Profile{
				Name: "normal profile",
			},
			exp: func(p Profile) bool {
				return p.BackgroundColor.RedComponent == 0 && p.BackgroundColor.GreenComponent == 0 && p.BackgroundColor.BlueComponent == 0
			},
		},
	}

	for _, test := range cases {
		test.profile.Colors()
		assert.True(t, test.exp(test.profile))
	}
}

func TestSmartSelectionRules(t *testing.T) {
	cases := []struct {
		name           string
		customContents string
		exp            func(rules []SmartSelectionRule) bool
	}{
		{
			name: "user rules",
			customContents: heredoc.Doc(`
				[
					{
					  "notes" : "jira ticket link",
					  "precision" : "very_high",
					  "regex" : "JENKINS-\\d*",
					  "actions" : [
						{
						  "title" : "Open Jenkins jira link",
						  "action" : 1,
						  "parameter" : "https://issues.jenkins-ci.org/browse/\\0"
						}
					  ]
					}
				]
			`),
			exp: func(rules []SmartSelectionRule) bool {
				return rules[len(rules)-1].Notes == "jira ticket link"
			},
		},
	}

	for _, test := range cases {
		file, cleanup := tempFile(t, test.customContents)
		defer cleanup()

		assert.True(t, test.exp(SmartSelectionRules(file)), test.name)

	}
}

func tempFile(t *testing.T, contents string) (string, func()) {
	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Fatal(err)
	}

	tmpfn := filepath.Join(dir, "tmpfile")
	if err := ioutil.WriteFile(tmpfn, []byte(contents), 0o666); err != nil {
		t.Fatal(err)
	}
	return tmpfn, func() {
		os.RemoveAll(dir)
	}
}

func TestProfileTree(t *testing.T) {
	cases := []struct {
		name     string
		profiles Profiles
		out      map[string][]string
	}{
		{
			name: "multiple child accounts",
			profiles: Profiles{
				Profiles: []Profile{
					{
						GUID: "parent",
					},
					{
						GUID: "child1",
						Tags: []string{
							"source-profile=parent",
						},
					},
					{
						GUID: "child2",
						Tags: []string{
							"source-profile=parent",
						},
					},
				},
			},
			out: map[string][]string{
				"parent": {
					"child1",
					"child2",
				},
			},
		},
	}

	for _, test := range cases {
		assert.Equal(t, test.profiles.ProfileTree(), test.out, test.name)
	}
}
