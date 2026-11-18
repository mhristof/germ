package ssm

import (
	"regexp"
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

// Command is a command sent to the remote shell when the SSM session of an
// instance whose name matches Match starts.
type Command struct {
	Match   string `mapstructure:"match"`
	Command string `mapstructure:"command"`
}

// ApplyCommands adds a trigger that sends the command of the first matching
// entry to the remote shell once the SSM session for an instance starts.
func ApplyCommands(profiles []iterm.Profile, commands []Command) {
	if len(commands) == 0 {
		return
	}

	type compiled struct {
		re      *regexp.Regexp
		command string
	}

	var rules []compiled

	for _, c := range commands {
		re, err := regexp.Compile(c.Match)
		if err != nil {
			log.Error().Err(err).Str("match", c.Match).Msg("invalid ssm command regex")
			continue
		}

		rules = append(rules, compiled{re: re, command: c.Command})
	}

	for i := range profiles {
		name, ok := instanceName(profiles[i].Name)
		if !ok {
			continue
		}

		for _, rule := range rules {
			if !rule.re.MatchString(name) {
				continue
			}

			profiles[i].Triggers = append(profiles[i].Triggers, iterm.Trigger{
				Action:    "SendTextTrigger",
				Parameter: rule.command + "\\r",
				Regex:     sessionStartRegex,
			})

			log.Debug().Str("profile", profiles[i].Name).Str("instance", name).Str("match", rule.re.String()).Msg("added ssm session command")

			break
		}
	}
}

// Commands returns the per-instance commands from the `ssm.commands` key of
// the germ config file.
func Commands() []Command {
	var commands []Command

	err := viper.UnmarshalKey("ssm.commands", &commands)
	if err != nil {
		log.Error().Err(err).Msg("cannot parse ssm.commands")
	}

	return commands
}
