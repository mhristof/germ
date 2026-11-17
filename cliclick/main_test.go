package cliclick

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProfile(t *testing.T) {
	cases := []struct {
		name        string
		setupBinary bool
		expectError bool
	}{
		{
			name:        "cliclick binary exists",
			setupBinary: true,
			expectError: false,
		},
		{
			name:        "cliclick binary missing",
			setupBinary: false,
			expectError: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			originalPath := os.Getenv("PATH")
			defer os.Setenv("PATH", originalPath)

			cliclickPath := ""
			if test.setupBinary {
				tempDir := t.TempDir()
				cliclickPath = filepath.Join(tempDir, "cliclick")
				err := os.WriteFile(cliclickPath, []byte("#!/bin/bash\necho 'mock cliclick'"), 0755)
				assert.NoError(t, err)

				os.Setenv("PATH", tempDir+":"+originalPath)
			} else {
				os.Setenv("PATH", "")
			}

			profile, err := Profile()

			if test.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot find cliclick binary")
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, "cliclick", profile.Name)
			assert.Equal(t, "Yes", profile.CustomCommand)
			assert.Equal(t,
				"/usr/bin/env CLICLICK="+cliclickPath+` /bin/sh -c 'while sleep 10; do "$CLICLICK" m:$("$CLICLICK" p); done'`,
				profile.Command,
			)

			keyMap, exists := profile.KeyboardMap["0x77-0x100000-0xd"]
			assert.True(t, exists)
			assert.Contains(t, keyMap.Text, "Cmd+w is disabled")
		})
	}
}
