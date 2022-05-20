package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kevinburke/ssh_config"

	"github.com/mhristof/germ/iterm"
	"github.com/rs/zerolog/log"
)

func containsPattern(patterns []*ssh_config.Pattern, needle string) bool {
	for _, pattern := range patterns {
		if strings.Contains(pattern.String(), needle) {
			return true
		}
	}

	return false
}

func Profiles() []iterm.Profile {
	config := filepath.Join(os.Getenv("HOME"), ".ssh/config")
	data, err := ioutil.ReadFile(config)
	if err != nil {
		log.Error().Str("config", config).Err(err).Msg("cannot open ssh config")
	}

	var ret []iterm.Profile

	for _, line := range strings.Split(string(data), "\n") {
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
		p := iterm.NewProfile(host, map[string]string{
			"Command": fmt.Sprintf("ssh %s", host),
		})

		p.Tags = []string{
			"ssh",
			hostIP(config, host),
		}

		ret = append(ret, *p)
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
