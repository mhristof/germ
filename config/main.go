package config

import (
	"path/filepath"
	"strings"

	"github.com/mhristof/germ/iterm"
	"github.com/mhristof/germ/paths"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func Path() string {
	return paths.Config()
}

func Load() {
	path := filepath.Dir(Path())
	name := strings.Split(filepath.Base(Path()), ".")
	viper.SetConfigName(name[0]) // name of config file (without extension)
	viper.SetConfigType(name[1]) // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(path)    // path to look for the config file in

	err := viper.ReadInConfig()
	if err != nil {
		log.Warn().Err(err).Msg("cannot parse config")
	}

	return
}

func Generate() []iterm.Profile {
	profiles := viper.GetStringMap("profiles")

	ret := make([]iterm.Profile, len(profiles))
	i := 0

	for profile := range viper.GetStringMap("profiles") {
		config := canonicalKeys(viper.GetStringMapString("profiles." + profile + ".config"))
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

// keys are the profile config keys iterm.NewProfile expects, in their case
// sensitive form.
var keys = []string{
	"AllowTitleSetting",
	"BadgeText",
	"Command",
	"Initial Text",
	"Tags",
}

// canonicalKeys restores the case of the known keys, as viper lowercases all
// the keys of the config file.
func canonicalKeys(config map[string]string) map[string]string {
	ret := make(map[string]string, len(config))
	for k, v := range config {
		ret[k] = v
	}

	for _, key := range keys {
		lower := strings.ToLower(key)
		if v, found := ret[lower]; found && lower != key {
			ret[key] = v
			delete(ret, lower)
		}
	}

	return ret
}
