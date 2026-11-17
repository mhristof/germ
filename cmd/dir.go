package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/mhristof/germ/keychain"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var clearDir bool

var setDirCmd = &cobra.Command{
	Use:     "set-dir [name [dir]]",
	Aliases: []string{"dir"},
	Short:   "Show or set the starting directory of a profile. Without a name, list the directories of all profiles",
	Args:    cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		Verbose(cmd)

		if len(args) == 0 {
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			for _, item := range keyChain.Items() {
				fmt.Fprintf(w, "%s\t%s\n", item.Account, item.Directory)
			}
			w.Flush()

			return
		}

		name := args[0]

		switch {
		case clearDir && len(args) == 2:
			log.Fatal().Msg("--clear does not take a directory")
		case clearDir:
			keyChain.SetDirectory(name, "")
			fmt.Printf("%s: directory cleared\n", name)
		case len(args) == 2:
			dir := absDir(args[1])
			keyChain.SetDirectory(name, dir)
			fmt.Printf("%s: directory set to %s\n", name, dir)
		default:
			item, found := findItem(name)
			if !found {
				log.Fatal().Str("name", name).Msg("keychain item does not exist")
			}

			if item.Directory == "" {
				fmt.Printf("%s: no directory set\n", name)
			} else {
				fmt.Printf("%s: %s\n", name, item.Directory)
			}
		}
	},
}

func findItem(name string) (keychain.Item, bool) {
	for _, item := range keyChain.Items() {
		if item.Account == name {
			return item, true
		}
	}

	return keychain.Item{}, false
}

// absDir expands `~` and turns the path into an absolute one, as iTerm
// resolves a relative working directory against its own cwd.
func absDir(path string) string {
	dir, err := filepath.Abs(expandUser(path))
	if err != nil {
		log.Fatal().Err(err).Str("path", path).Msg("cannot resolve directory")
	}

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		log.Warn().Str("dir", dir).Msg("directory does not exist")
	}

	return dir
}

func init() {
	setDirCmd.Flags().BoolVarP(&clearDir, "clear", "", false, "Clear the starting directory")

	rootCmd.AddCommand(setDirCmd)
}
