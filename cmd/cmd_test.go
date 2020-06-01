package cmd

import (
	"testing"

	"github.com/mhristof/germ/iterm"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTemplate(t *testing.T) {
	var cases = []struct {
		name    string
		command string
		profile string
		out     []string
	}{
		{
			name:    "simple command with out templates",
			command: "aws s3 ls",
			profile: "foo",
			out:     []string{"aws s3 ls"},
		},
		{
			name:    "template the profile name",
			command: "aws s3 ls > {{ .Profile }}",
			profile: "foo",
			out:     []string{"aws s3 ls > foo"},
		},
		{
			name:    "template with the region",
			command: "aws s3 ls --region {{ .Region }}",
			profile: "foo",
			out: []string{
				"aws s3 ls --region us-east-2",
				"aws s3 ls --region us-east-1",
				"aws s3 ls --region us-west-1",
				"aws s3 ls --region us-west-2",
				"aws s3 ls --region af-south-1",
				"aws s3 ls --region ap-east-1",
				"aws s3 ls --region ap-south-1",
				"aws s3 ls --region ap-northeast-3",
				"aws s3 ls --region ap-northeast-2",
				"aws s3 ls --region ap-southeast-1",
				"aws s3 ls --region ap-southeast-2",
				"aws s3 ls --region ap-northeast-1",
				"aws s3 ls --region ca-central-1",
				"aws s3 ls --region cn-north-1",
				"aws s3 ls --region cn-northwest-1",
				"aws s3 ls --region eu-central-1",
				"aws s3 ls --region eu-west-1",
				"aws s3 ls --region eu-west-2",
				"aws s3 ls --region eu-south-1",
				"aws s3 ls --region eu-west-3",
				"aws s3 ls --region eu-north-1",
				"aws s3 ls --region me-south-1",
				"aws s3 ls --region sa-east-1",
			},
		},
	}

	for _, test := range cases {
		assert.Equal(t, generateTemplate(test.command, test.profile), test.out, test.name)
	}
}

func TestGenerateCommands(t *testing.T) {
	var cases = []struct {
		name     string
		profiles iterm.Profiles
		command  string
		out      []string
	}{
		{
			name:    "2 profiles, one is the login for the other",
			command: "aws s3 ls",
			profiles: iterm.Profiles{
				Profiles: []iterm.Profile{
					iterm.Profile{
						GUID: "parent",
					},
					iterm.Profile{
						GUID:    "login-parent",
						Command: "login-command",
					},
					iterm.Profile{
						GUID: "child",
						Tags: []string{
							"source-profile=parent",
						},
					},
				},
			},
			out: []string{
				"login-command",
				"AWS_PROFILE=child aws s3 ls",
			},
		},
		{
			name:    "test that the login command is generate once",
			command: "aws s3 ls",
			profiles: iterm.Profiles{
				Profiles: []iterm.Profile{
					iterm.Profile{
						GUID: "parent",
					},
					iterm.Profile{
						GUID:    "login-parent",
						Command: "login-command",
					},
					iterm.Profile{
						GUID: "child1",
						Tags: []string{
							"source-profile=parent",
						},
					},
					iterm.Profile{
						GUID: "child2",
						Tags: []string{
							"source-profile=parent",
						},
					},
				},
			},
			out: []string{
				"login-command",
				"AWS_PROFILE=child1 aws s3 ls",
				"AWS_PROFILE=child2 aws s3 ls",
			},
		},
		{
			name:    "login command with sleep at the end",
			command: "aws s3 ls",
			profiles: iterm.Profiles{
				Profiles: []iterm.Profile{
					iterm.Profile{
						GUID: "parent",
					},
					iterm.Profile{
						GUID:    "login-parent",
						Command: "bash -c 'login-command || sleep 60'",
					},
					iterm.Profile{
						GUID: "child",
						Tags: []string{
							"source-profile=parent",
						},
					},
				},
			},
			out: []string{
				"bash -c 'login-command'",
				"AWS_PROFILE=child aws s3 ls",
			},
		},
	}

	for _, test := range cases {
		assert.Equal(t, generateCommands(test.profiles, test.command), test.out, test.name)

	}
}
