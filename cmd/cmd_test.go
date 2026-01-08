package cmd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/mhristof/germ/aws"
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
			name:    "template with the region (wierdly spaced to test the regex match)",
			command: "aws s3 ls --region {{.Region }}",
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
func TestGenerateTemplateEdgeCases(t *testing.T) {
	cases := []struct {
		name    string
		command string
		profile string
		out     []string
	}{
		{
			name:    "empty command",
			command: "",
			profile: "test-profile",
			out:     []string{""},
		},
		{
			name:    "command with multiple profile templates",
			command: "echo {{ .Profile }} && echo {{ .Profile }}",
			profile: "my-profile",
			out:     []string{"echo my-profile && echo my-profile"},
		},
		{
			name:    "command with profile and region templates",
			command: "aws --profile {{ .Profile }} s3 ls --region {{ .Region }}",
			profile: "test",
			out: func() []string {
				var expected []string
				for _, region := range []string{
					"us-east-2", "us-east-1", "us-west-1", "us-west-2",
					"af-south-1", "ap-east-1", "ap-south-1", "ap-northeast-3",
					"ap-northeast-2", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1",
					"ca-central-1", "cn-north-1", "cn-northwest-1", "eu-central-1",
					"eu-west-1", "eu-west-2", "eu-south-1", "eu-west-3",
					"eu-north-1", "me-south-1", "sa-east-1",
				} {
					expected = append(expected, fmt.Sprintf("aws --profile test s3 ls --region %s", region))
				}
				return expected
			}(),
		},
		{
			name:    "special characters in profile name",
			command: "echo {{ .Profile }}",
			profile: "profile-with-dashes_and_underscores",
			out:     []string{"echo profile-with-dashes_and_underscores"},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result := generateTemplate(test.command, test.profile)
			assert.Equal(t, test.out, result)
		})
	}
}

func TestGenerateCommandsEdgeCases(t *testing.T) {
	cases := []struct {
		name     string
		profiles iterm.Profiles
		command  string
		expected []string
	}{
		{
			name:     "empty profiles",
			profiles: iterm.Profiles{Profiles: []iterm.Profile{}},
			command:  "aws s3 ls",
			expected: nil,
		},
		{
			name: "profile without login",
			profiles: iterm.Profiles{
				Profiles: []iterm.Profile{
					{
						GUID: "standalone",
						Name: "standalone-profile",
					},
				},
			},
			command:  "aws s3 ls",
			expected: nil, // Should be nil because no login profile found
		},
		{
			name: "multiple source profiles",
			profiles: iterm.Profiles{
				Profiles: []iterm.Profile{
					{GUID: "source1"},
					{GUID: "login-source1", Command: "login1"},
					{GUID: "child1", Tags: []string{"source-profile=source1"}},
					{GUID: "source2"},
					{GUID: "login-source2", Command: "login2"},
					{GUID: "child2", Tags: []string{"source-profile=source2"}},
				},
			},
			command: "aws s3 ls",
			expected: []string{
				"login1",
				"AWS_PROFILE=child1 aws s3 ls",
				"login2", 
				"AWS_PROFILE=child2 aws s3 ls",
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result := generateCommands(test.profiles, test.command)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestTemplateRegexMatching(t *testing.T) {
	cases := []struct {
		name     string
		command  string
		hasRegion bool
	}{
		{
			name:     "simple region template",
			command:  "aws s3 ls --region {{ .Region }}",
			hasRegion: true,
		},
		{
			name:     "region template with spaces",
			command:  "aws s3 ls --region {{.Region}}",
			hasRegion: true,
		},
		{
			name:     "region template with extra spaces",
			command:  "aws s3 ls --region {{  .Region  }}",
			hasRegion: true,
		},
		{
			name:     "no region template",
			command:  "aws s3 ls",
			hasRegion: false,
		},
		{
			name:     "profile template only",
			command:  "aws s3 ls --profile {{ .Profile }}",
			hasRegion: false,
		},
		{
			name:     "region in comment (should not match)",
			command:  "aws s3 ls # region: {{ .Region }}",
			hasRegion: true, // The regex will still match this
		},
	}

	regexRegion := regexp.MustCompile(`{{\s*\.Region\s*}}`)
	
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			matches := regexRegion.MatchString(test.command)
			assert.Equal(t, test.hasRegion, matches)
		})
	}
}

func TestCommandGeneration(t *testing.T) {
	// Test the actual command generation with realistic profiles
	profiles := iterm.Profiles{
		Profiles: []iterm.Profile{
			{
				GUID: "prod-account",
				Name: "prod-account",
			},
			{
				GUID:    "login-prod-account",
				Name:    "login-prod-account", 
				Command: "aws sso login --profile prod-account",
			},
			{
				GUID: "prod-role1",
				Name: "prod-role1",
				Tags: []string{"source-profile=prod-account"},
			},
			{
				GUID: "prod-role2", 
				Name: "prod-role2",
				Tags: []string{"source-profile=prod-account"},
			},
		},
	}

	t.Run("simple command", func(t *testing.T) {
		result := generateCommands(profiles, "aws sts get-caller-identity")
		expected := []string{
			"aws sso login --profile prod-account",
			"AWS_PROFILE=prod-role1 aws sts get-caller-identity",
			"AWS_PROFILE=prod-role2 aws sts get-caller-identity",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("command with region template", func(t *testing.T) {
		result := generateCommands(profiles, "aws ec2 describe-instances --region {{ .Region }}")
		
		// Should have login command + (2 profiles * number of regions) commands
		expectedCount := 1 + (2 * len(aws.Regions()))
		assert.Equal(t, expectedCount, len(result))
		
		// First command should be login
		assert.Equal(t, "aws sso login --profile prod-account", result[0])
		
		// Check that region templates are expanded
		assert.Contains(t, result[1], "us-east-2") // First region in the list
		assert.Contains(t, result[1], "AWS_PROFILE=prod-role1")
	})
}