package cmd

import (
	"fmt"
	"runtime"

	"github.com/mhristof/go-update"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the binary with a new version",
	Run: func(cmd *cobra.Command, args []string) {
		url := fmt.Sprintf("https://github.com/mhristof/germ/releases/latest/download/germ.%s", runtime.GOOS)

		updates, updateFunc, err := update.Check(url)
		if err != nil {
			panic(err)
		}

		if !updates {
			return
		}

		if silent, _ := cmd.Flags().GetBool("silent"); !silent {
			log.Info().Bool("updates", updates).Msg("updates avalable")
		}

		if dryrun, _ := cmd.Flags().GetBool("dryrun"); dryrun {
			return
		}

		log.Info().Str("url", url).Msg("downloading new version")

		updateFunc()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
