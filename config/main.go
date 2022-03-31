package config

import (
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/mhristof/germ/iterm"
	"github.com/mhristof/germ/log"
	"github.com/spf13/viper"
)

func Path() string {
	return filepath.Join(xdg.ConfigHome, "germ.yaml")
}

func Load() {
	path := filepath.Dir(Path())
	name := strings.Split(filepath.Base(Path()), ".")
	viper.SetConfigName(name[0]) // name of config file (without extension)
	viper.SetConfigType(name[1]) // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(path)    // path to look for the config file in

	err := viper.ReadInConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Warning("cannot parse config")
	}

	return
}

func Generate() []iterm.Profile {
	profiles := viper.GetStringMap("profiles")

	ret := make([]iterm.Profile, len(profiles))
	i := 0

	for profile := range viper.GetStringMap("profiles") {
		config := viper.GetStringMapString("profiles." + profile + ".config")
		pro := iterm.NewProfile(profile, config)

		var triggers []iterm.Trigger

		err := viper.UnmarshalKey("profiles."+profile+".triggers", &triggers)
		if err == nil {
			pro.Triggers = append(pro.Triggers, triggers...)
		}

		ret[i] = *pro
		i++
	}

	return ret
}
