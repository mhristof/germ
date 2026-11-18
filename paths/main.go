// Package paths holds the locations of the files germ reads and writes,
// following the XDG base directory spec.
package paths

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

const app = "germ"

// base returns the XDG base directory from env, or the spec default under
// $HOME. The macOS ~/Library defaults are not used on purpose, to keep the
// files with the rest of the dotfiles. The env is read on every call so tests
// can redirect it.
func base(env, def string) string {
	if dir := os.Getenv(env); filepath.IsAbs(dir) {
		return dir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Err(err).Msg("cannot find home directory")
	}

	return filepath.Join(home, def)
}

func configHome() string { return base("XDG_CONFIG_HOME", ".config") }
func stateHome() string  { return base("XDG_STATE_HOME", filepath.Join(".local", "state")) }
func cacheHome() string  { return base("XDG_CACHE_HOME", ".cache") }

// Config is the user edited config file.
func Config() string {
	return filepath.Join(configHome(), app, "config.yaml")
}

// SSR is the user edited smart selection rules file.
func SSR() string {
	return filepath.Join(configHome(), app, "ssr.json")
}

// Names is the generated file that keeps the profile badge names stable.
func Names() string {
	return filepath.Join(stateHome(), app, "profiles.json")
}

// Cache is a generated file that can be recreated at any time.
func Cache(name string) string {
	return filepath.Join(cacheHome(), app, name)
}

// legacy maps the old file locations to the new ones.
func legacy() map[string]string {
	home, _ := os.UserHomeDir()

	return map[string]string{
		filepath.Join(configHome(), "germ.yaml"):          Config(),
		filepath.Join(home, ".germ.ssr.json"):             SSR(),
		filepath.Join(configHome(), "germ-profiles.json"): Names(),
		filepath.Join(cacheHome(), "germ.ssm.json"):       Cache("ssm.json"),
	}
}

// Migrate moves files from their old locations, unless the new file exists.
func Migrate() {
	for old, cur := range legacy() {
		migrate(old, cur)
	}
}

func migrate(old, cur string) {
	if _, err := os.Stat(old); err != nil {
		return
	}

	if _, err := os.Stat(cur); err == nil {
		log.Warn().Str("old", old).Str("new", cur).Msg("both old and new files exist, ignoring the old one")
		return
	}

	err := os.MkdirAll(filepath.Dir(cur), 0o755)
	if err == nil {
		err = os.Rename(old, cur)
	}

	if err != nil {
		log.Error().Err(err).Str("old", old).Str("new", cur).Msg("cannot migrate file")
		return
	}

	log.Info().Str("old", old).Str("new", cur).Msg("migrated file")
}

// Write writes data to path, creating the parent directories.
func Write(path string, data []byte) error {
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
