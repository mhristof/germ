package vault

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/mhristof/germ/iterm"
	"github.com/stretchr/testify/assert"
)

func TestProfile(t *testing.T) {
	cases := []struct {
		name        string
		setupVault  bool
		expectError bool
	}{
		{
			name:        "vault binary exists",
			setupVault:  true,
			expectError: false,
		},
		{
			name:        "vault binary missing",
			setupVault:  false,
			expectError: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			originalPath := os.Getenv("PATH")
			
			if test.setupVault {
				// Create a mock vault binary
				tempDir := t.TempDir()
				vaultPath := filepath.Join(tempDir, "vault")
				err := os.WriteFile(vaultPath, []byte("#!/bin/bash\necho 'mock vault'"), 0755)
				assert.NoError(t, err)
				
				// Add temp dir to PATH
				os.Setenv("PATH", tempDir+":"+originalPath)
			} else {
				// Set empty PATH to simulate missing vault
				os.Setenv("PATH", "")
			}
			
			defer os.Setenv("PATH", originalPath)

			profile, err := Profile()

			if test.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot find vault binary")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "vault", profile.Name)
				assert.NotEmpty(t, profile.GUID)
				assert.Contains(t, profile.Command, "vault server -dev")
				
				// Check triggers
				assert.Len(t, profile.Triggers, 1)
				trigger := profile.Triggers[0]
				assert.Equal(t, "CoprocessTrigger", trigger.Action)
				assert.Equal(t, `echo '\1' > ~/.vault-token`, trigger.Parameter)
				assert.Equal(t, "^Root Token: (s.*)", trigger.Regex)
				assert.True(t, trigger.Partial)
				
				// Check bound hosts
				assert.Equal(t, []string{"&vault"}, profile.BoundHosts)
				
				// Check keyboard mapping
				keyMap, exists := profile.KeyboardMap["0x77-0x100000-0xd"]
				assert.True(t, exists)
				assert.Equal(t, int64(1), keyMap.Version)
				assert.Equal(t, int64(12), keyMap.Action)
				assert.Contains(t, keyMap.Text, "Cmd+w is disabled")
			}
		})
	}
}

func TestVaultBinaryLookup(t *testing.T) {
	// Test the actual binary lookup logic
	originalPath := os.Getenv("PATH")
	defer os.Setenv("PATH", originalPath)

	t.Run("binary exists in PATH", func(t *testing.T) {
		// Create a temporary directory with a mock vault binary
		tempDir := t.TempDir()
		vaultPath := filepath.Join(tempDir, "vault")
		err := os.WriteFile(vaultPath, []byte("#!/bin/bash\necho 'vault'"), 0755)
		assert.NoError(t, err)

		os.Setenv("PATH", tempDir)

		path, err := exec.LookPath("vault")
		assert.NoError(t, err)
		assert.Equal(t, vaultPath, path)
	})

	t.Run("binary does not exist", func(t *testing.T) {
		os.Setenv("PATH", "")

		_, err := exec.LookPath("vault")
		assert.Error(t, err)
	})
}

func TestProfileStructure(t *testing.T) {
	// Create a mock vault binary for this test
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault")
	err := os.WriteFile(vaultPath, []byte("#!/bin/bash\necho 'vault'"), 0755)
	assert.NoError(t, err)

	originalPath := os.Getenv("PATH")
	os.Setenv("PATH", tempDir+":"+originalPath)
	defer os.Setenv("PATH", originalPath)

	profile, err := Profile()
	assert.NoError(t, err)

	// Verify it's a valid iterm.Profile
	assert.IsType(t, iterm.Profile{}, profile)
	assert.NotEmpty(t, profile.Name)
	assert.NotEmpty(t, profile.GUID)
	assert.NotNil(t, profile.Triggers)
	assert.NotNil(t, profile.BoundHosts)
	assert.NotNil(t, profile.KeyboardMap)
}