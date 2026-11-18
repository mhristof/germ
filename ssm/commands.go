package ssm

import (
	"strings"

	"github.com/mhristof/germ/iterm"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// sessionStartRegex matches the line printed by `aws ssm start-session` once
// the session is established, so the trigger fires once per connection.
const sessionStartRegex = "^Starting session with SessionId: "

// instanceName extracts the instance name from a profile named
// <account>:<region>:ssm-<instance>.
func instanceName(profileName string) (string, bool) {
	parts := strings.SplitN(profileName, ":", 3)
	if len(parts) != 3 || !strings.HasPrefix(parts[2], "ssm-") {
		return "", false
	}

	return strings.TrimPrefix(parts[2], "ssm-"), true
}

// ApplyCommands adds a trigger that sends the configured command to the
// remote shell once the SSM session for a matching instance starts.
// Instance names are matched case insensitively, as viper lowercases keys.
func ApplyCommands(profiles []iterm.Profile, commands map[string]string) {
	if len(commands) == 0 {
		return
	}

	for i := range profiles {
		name, ok := instanceName(profiles[i].Name)
		if !ok {
			continue
		}

		command, found := commands[strings.ToLower(name)]
		if !found {
			continue
		}

		profiles[i].Triggers = append(profiles[i].Triggers, iterm.Trigger{
			Action:    "SendTextTrigger",
			Parameter: command + "\\r",
			Regex:     sessionStartRegex,
		})

		log.Debug().Str("profile", profiles[i].Name).Str("instance", name).Msg("added ssm session command")
	}
}

// Commands returns the per-instance commands from the `ssm.commands` key of
// the germ config file.
func Commands() map[string]string {
	return viper.GetStringMapString("ssm.commands")
}
