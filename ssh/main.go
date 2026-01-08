package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mhristof/germ/iterm"
	"github.com/mhristof/germ/profile"
	"github.com/rs/zerolog/log"
)

func Profiles() []iterm.Profile {
	config := filepath.Join(os.Getenv("HOME"), ".ssh/config")
	data, err := ioutil.ReadFile(config)
	if err != nil {
		log.Error().Str("config", config).Err(err).Msg("cannot open ssh config")
	}

	var ret []iterm.Profile
	var lastProfile *iterm.Profile

	for _, line := range strings.Split(string(data), "\n") {
		// Handle tmux configuration for the last created profile
		if strings.Contains(line, "RemoteCommand tmux") && lastProfile != nil {
			// Update the last profile with tmux detach shortcut
			updatedProfile := profile.NewSSHProfileBuilder(lastProfile.Name).
				WithSSHCommand(lastProfile.Name).
				WithTmuxDetach().
				Build()
			
			// Replace the last profile with the updated one
			ret[len(ret)-1] = *updatedProfile
			
			log.Debug().Str("line", line).Str("profile", lastProfile.Name).Msg("found tmux for profile")
		}

		if !strings.HasPrefix(line, "Host ") {
			continue
		}

		if strings.Contains(line, "*") {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) != 2 {
			log.Debug().Interface("fields", fields).Msg("more fields than expected")
			continue
		}

		host := fields[1]
		hostIPAddr := hostIP(config, host)
		
		sshProfile := profile.NewSSHProfileBuilder(host).
			WithSSHCommand(host).
			WithHostIP(hostIPAddr).
			Build()

		ret = append(ret, *sshProfile)
		lastProfile = sshProfile
	}

	return ret
}

func hostIP(config, host string) string {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("ssh -G %s", host))

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Warn().Str("host", host).
			Str("config", config).
			Str("stderr.String()", stderr.String()).
			Msg("cannot find IP for host")

		return ""
	}

	for _, line := range strings.Split(stdout.String(), "\n") {
		if !strings.HasPrefix(line, "hostname ") {
			continue
		}

		return strings.Fields(line)[1]
	}

	return ""
}
