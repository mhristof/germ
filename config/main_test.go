package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestPath(t *testing.T) {
	path := Path()
	assert.Contains(t, path, "germ.yaml")
	assert.True(t, filepath.IsAbs(path))
}

func TestGenerate(t *testing.T) {
	cases := []struct {
		name     string
		config   map[string]interface{}
		expected int
	}{
		{
			name: "single profile",
			config: map[string]interface{}{
				"profiles": map[string]interface{}{
					"test-profile": map[string]interface{}{
						"config": map[string]interface{}{
							"Command": "echo test",
							"Tags":    "test,profile",
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "multiple profiles",
			config: map[string]interface{}{
				"profiles": map[string]interface{}{
					"profile1": map[string]interface{}{
						"config": map[string]interface{}{
							"Command": "echo profile1",
						},
					},
					"profile2": map[string]interface{}{
						"config": map[string]interface{}{
							"Command": "echo profile2",
						},
					},
				},
			},
			expected: 2,
		},
		{
			name: "profile with triggers",
			config: map[string]interface{}{
				"profiles": map[string]interface{}{
					"trigger-profile": map[string]interface{}{
						"config": map[string]interface{}{
							"Command": "echo trigger",
						},
						"triggers": []map[string]interface{}{
							{
								"regex":  "error",
								"action": "highlight",
							},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name:     "no profiles",
			config:   map[string]interface{}{},
			expected: 0,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			// Reset viper for each test
			viper.Reset()
			
			// Set test configuration
			for key, value := range test.config {
				viper.Set(key, value)
			}

			profiles := Generate()
			assert.Equal(t, test.expected, len(profiles))

			if test.expected > 0 {
				// Verify profile structure
				for _, profile := range profiles {
					assert.NotEmpty(t, profile.Name)
					assert.NotEmpty(t, profile.GUID)
				}

				// Test specific cases
				if test.name == "single profile" {
					assert.Equal(t, "test-profile", profiles[0].Name)
				}

				if test.name == "profile with triggers" {
					assert.Equal(t, "trigger-profile", profiles[0].Name)
					// Note: Triggers testing would require more complex setup
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "germ.yaml")
	
	configContent := `profiles:
  test-profile:
    config:
      Command: "echo test"
      Tags: "test"
`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Mock the config path by setting XDG_CONFIG_HOME
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tempDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalConfigHome)

	// Reset viper
	viper.Reset()

	// Test Load function - it should not panic even if config doesn't exist in expected location
	assert.NotPanics(t, func() {
		Load()
	})
}

func TestLoadNonExistentConfig(t *testing.T) {
	// Set a non-existent config directory
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	
	originalConfigHome := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", nonExistentDir)
	defer os.Setenv("XDG_CONFIG_HOME", originalConfigHome)

	// Reset viper
	viper.Reset()

	// This should not panic, just log a warning
	assert.NotPanics(t, func() {
		Load()
	})
}