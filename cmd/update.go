package cmd

import (
	"fmt"
	"runtime"

	"github.com/mhristof/germ/log"
	"github.com/mhristof/go-update"
	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
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
				log.WithFields(log.Fields{
					"updates": updates,
				}).Info("Updates available")
			}

			if dryrun, _ := cmd.Flags().GetBool("dryrun"); dryrun {
				return
			}

			log.WithFields(log.Fields{
				"url": url,
			}).Info("Downloading new version")

			updateFunc()
		},
	}
)

func init() {
	rootCmd.AddCommand(updateCmd)
}
