package aws

import (
	"fmt"
	"testing"

	"github.com/mhristof/gterm/iterm"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	var cases = []struct {
		name     string
		config   []map[string]string
		expected []iterm.Profile
	}{
		{
			name: "single profile",
			config: []map[string]string{
				{
					"foo": "bar",
				},
			},
			expected: []iterm.Profile{
				{
					GUID: "0",
				},
				{
					GUID: "login-0",
				},
			},
		},
		{
			name: "profile with source_profile",
			config: []map[string]string{
				{
					"source_profile": "bar",
				},
			},
			expected: []iterm.Profile{
				{
					GUID: "0",
				},
			},
		},
	}

	for _, test := range cases {
		var prof iterm.Profiles
		for i, cfg := range test.config {
			add(&prof, fmt.Sprintf("%d", i), cfg)
		}

		assert.Equal(t, len(test.expected), len(prof.Profiles))

		for _, profile := range test.expected {
			_, found := prof.FindGUID(profile.GUID)
			assert.True(t, found)
		}

	}
}
