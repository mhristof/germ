package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/mhristof/germ/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	dryRun  bool
	version = "devel"
)

var rootCmd = &cobra.Command{
	Use:   "germ",
	Short: "Generate dynamic iTerm2 profiles",
	Long: heredoc.Doc(fmt.Sprintf(`
		To add your custom profiles, you can create a file in
		"%s"
		with the following contents:

		---
		profiles:
		  name:
		    config:
		      command: "date"
	`, config.Path())),
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		Verbose(cmd)
	},
}

func Verbose(cmd *cobra.Command) {
	verbose, err := cmd.Flags().GetCount("verbose")
	if err != nil {
		panic(err)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})

	switch verbose {
	case 1:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case 2:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func init() {
	rootCmd.PersistentFlags().CountP("verbose", "v", "Increase verbosity")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dryrun", "n", false, "Dry run mode, no changes will be made on the system")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Panic().Err(err).Msg("cannot execute command")
	}
}
