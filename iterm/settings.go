package iterm

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

// ConfigureGlobalSettings ensures iTerm global settings are configured correctly.
// Returns a list of settings that were changed.
func ConfigureGlobalSettings() []string {
	var changed []string

	// Tab bar at bottom (TabViewType: 0=top, 1=bottom, 2=left)
	if configureDefault("TabViewType", "1", "-int") {
		changed = append(changed, "Tab bar moved to bottom")
	}

	// Always show tab bar (HideTab: false = always show)
	if configureDefault("HideTab", "0", "-bool") {
		changed = append(changed, "Tab bar set to always visible")
	}

	return changed
}

// configureDefault sets an iTerm2 default if it differs from the desired value.
// Returns true if the value was changed.
func configureDefault(key, value, valueType string) bool {
	current := readDefault(key)
	
	// Normalize boolean values for comparison
	normalizedCurrent := normalizeBoolValue(current)
	normalizedValue := normalizeBoolValue(value)
	
	if normalizedCurrent == normalizedValue {
		log.Debug().Str("key", key).Str("value", value).Msg("iTerm setting already configured")
		return false
	}

	cmd := exec.Command("defaults", "write", "com.googlecode.iterm2", key, valueType, value)
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Str("key", key).Str("value", value).Msg("failed to set iTerm default")
		return false
	}

	log.Info().Str("key", key).Str("old", current).Str("new", value).Msg("iTerm setting updated")
	return true
}

// readDefault reads an iTerm2 default value.
func readDefault(key string) string {
	cmd := exec.Command("defaults", "read", "com.googlecode.iterm2", key)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// normalizeBoolValue converts various boolean representations to a standard form.
func normalizeBoolValue(val string) string {
	switch strings.ToLower(val) {
	case "0", "false", "no":
		return "0"
	case "1", "true", "yes":
		return "1"
	default:
		return val
	}
}

// PrintSettingsChanges prints the changes made to iTerm settings.
func PrintSettingsChanges(changes []string) {
	if len(changes) == 0 {
		return
	}

	fmt.Println("iTerm settings updated:")
	for _, change := range changes {
		fmt.Printf("  - %s\n", change)
	}
	fmt.Println("Restart iTerm for changes to take effect.")
}
