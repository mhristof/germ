package aws

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/mhristof/germ/iterm"
	"github.com/mhristof/germ/log"
	"github.com/zieckey/goini"
)

func Profiles(config string) []iterm.Profile {
	ini := goini.New()
	err := ini.ParseFile(config)
	if err != nil {
		log.WithFields(log.Fields{
			"config": config,
			"err":    err.Error(),
		}).Error("paarseINI file failed.")
		return nil
	}

	var prof iterm.Profiles
	for name, section := range ini.GetAll() {
		if name == "" {
			continue
		}
		tName := strings.TrimPrefix(name, "profile ")
		add(&prof, fmt.Sprintf("%s", tName), section)
	}

	return prof.Profiles
}

func add(p *iterm.Profiles, name string, config map[string]string) {
	user, err := user.Current()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Cannot find current user")
	}

	config["Command"] = fmt.Sprintf("/usr/bin/env AWS_PROFILE=%s /usr/bin/login -fp %s", name, user.Username)
	profile := iterm.NewProfile(name, config)
	p.Add(*profile)

	if _, found := config["source_profile"]; !found {
		config["Command"] = loginCmd(name, config)
		loginProfile := iterm.NewProfile(fmt.Sprintf("login-%s", name), config)
		p.Add(*loginProfile)
	}
}

func loginCmd(name string, config map[string]string) string {
	var tool, toolCmd string
	_, azure := config["azure_tenant_id"]

	if azure {
		tool = "aws-azure-login"
		toolCmd = fmt.Sprintf("%s --no-prompt", tool)
	} else {
		return ""
	}

	bin, err := exec.LookPath(tool)
	if err != nil {
		log.WithFields(log.Fields{
			"tool": tool,
			"err":  err,
		}).Fatal("Cannot find executable")
	}

	return fmt.Sprintf(
		"bash -c 'AWS_PROFILE=%s PATH=%s NODE_EXTRA_CA_CERTS=%s %s || sleep 60'",
		name, filepath.Dir(bin), os.Getenv("NODE_EXTRA_CA_CERTS"), toolCmd,
	)

}
