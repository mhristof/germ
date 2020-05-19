package cmd

import (
	"os"

	"github.com/mhristof/germ/log"
	"github.com/spf13/cobra"
)

var (
	dryRun bool
)

var rootCmd = &cobra.Command{
	Use:   "germ",
	Short: "Generate dynamic iTerm2 profiles",
	Run: func(cmd *cobra.Command, args []string) {
		Verbose(cmd)
	},
}

func Verbose(cmd *cobra.Command) {
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Panic("Cannot get verbose value")
	}

	if verbose {
		log.SetLevel(log.DebugLevel)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dryrun", "n", false, "Dry run mode, no changes will be made on the system")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Increase verbosity")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Unable to execute command")
		os.Exit(1)
	}
}
