package cmd

import (
	"testing"

	"github.com/mhristof/germ/iterm"
	"github.com/stretchr/testify/assert"
)

func TestBuildMatch(t *testing.T) {
	cases := []struct {
		name     string
		profile  iterm.Profile
		expected string
	}{
		{
			name: "profile with region tags",
			profile: iterm.Profile{
				Name: "my-prod-profile",
				Tags: []string{"Europe", "London", "euw2", "account=123456789012"},
			},
			expected: "my-prod-profile Europe London euw2 account=123456789012",
		},
		{
			name: "profile with no tags",
			profile: iterm.Profile{
				Name: "simple-profile",
				Tags: []string{},
			},
			expected: "simple-profile",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, buildMatch(test.profile))
		})
	}
}

func TestBuildSubtitle(t *testing.T) {
	cases := []struct {
		name     string
		profile  iterm.Profile
		expected string
	}{
		{
			name: "profile with account tag",
			profile: iterm.Profile{
				Name: "my-profile",
				Tags: []string{"account=123456789012", "Europe", "London"},
			},
			expected: "account=123456789012",
		},
		{
			name: "profile with no relevant tags",
			profile: iterm.Profile{
				Name: "my-profile",
				Tags: []string{"Europe", "London"},
			},
			expected: "AWS profile",
		},
		{
			name: "profile with source-profile tag",
			profile: iterm.Profile{
				Name: "my-profile",
				Tags: []string{"source-profile=base", "Europe"},
			},
			expected: "source-profile=base",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, buildSubtitle(test.profile))
		})
	}
}
