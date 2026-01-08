package vim

import (
	"strings"
	"testing"

	"github.com/mhristof/germ/iterm"
	"github.com/stretchr/testify/assert"
)

func TestProfile(t *testing.T) {
	profile := Profile()

	// Test basic profile properties
	assert.Equal(t, "vim", profile.Name)
	assert.NotEmpty(t, profile.GUID)
	assert.IsType(t, iterm.Profile{}, profile)

	// Test that config is initialized (even if empty)
	// Note: Profile struct doesn't have a Config field, but has individual fields

	// Test triggers are initialized as empty slice
	assert.NotNil(t, profile.Triggers)
	assert.Len(t, profile.Triggers, 0)

	// Test bound hosts
	expectedHosts := []string{"&vim", "&nvim"}
	assert.Equal(t, expectedHosts, profile.BoundHosts)
	assert.Len(t, profile.BoundHosts, 2)
	assert.Contains(t, profile.BoundHosts, "&vim")
	assert.Contains(t, profile.BoundHosts, "&nvim")
}

func TestProfileConsistency(t *testing.T) {
	// Test that multiple calls return consistent results
	profile1 := Profile()
	profile2 := Profile()

	assert.Equal(t, profile1.Name, profile2.Name)
	assert.Equal(t, profile1.BoundHosts, profile2.BoundHosts)
	assert.Equal(t, len(profile1.Triggers), len(profile2.Triggers))
	
	// GUIDs should be the same since they're based on the same name "vim"
	assert.Equal(t, profile1.GUID, profile2.GUID)
}

func TestProfileIntegration(t *testing.T) {
	profile := Profile()

	// Test that the profile can be used in a slice of profiles
	profiles := []iterm.Profile{profile}
	assert.Len(t, profiles, 1)
	assert.Equal(t, "vim", profiles[0].Name)

	// Test that bound hosts work as expected for vim/nvim detection
	for _, host := range profile.BoundHosts {
		assert.True(t, host == "&vim" || host == "&nvim")
		assert.True(t, strings.HasPrefix(host, "&"))
	}
}