package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"text/template"

	"github.com/MakeNowJust/heredoc"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	defaultProfileName   string
	defaultProfilePython = heredoc.Doc(`
		#!/usr/bin/env python3.7

		import iterm2

		async def main(connection):
			all_profiles = await iterm2.PartialProfile.async_query(connection)
			for profile in all_profiles:
				if profile.name == "{{ .Profile }}":
					await profile.async_make_default()
					return

		iterm2.run_until_complete(main)
	`)
)

var defaultCmd = &cobra.Command{
	Use:     "default",
	Aliases: []string{"def"},
	Short:   "Set the default profile in iterm",
	Run: func(cmd *cobra.Command, args []string) {
		tmpl, err := template.New("script").Parse(defaultProfilePython)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot create template")
		}

		rendered := new(bytes.Buffer)
		err = tmpl.Execute(rendered, struct {
			Profile string
		}{
			Profile: defaultProfileName,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("cannot render template")
		}

		tmpfile, err := ioutil.TempFile("", "default-profile")
		if err != nil {
			log.Fatal().Err(err).Msg("cannot create temp file")
		}
		defer os.Remove(tmpfile.Name())

		fmt.Println(tmpfile.Name())
		if _, err := tmpfile.Write(rendered.Bytes()); err != nil {
			log.Fatal().Err(err).Msg("cannot write file")
		}
		if err := tmpfile.Close(); err != nil {
			log.Fatal().Err(err).Msg("cannot close file")
		}

		python3, err := exec.LookPath("python3")
		if err != nil {
			log.Fatal().Err(err).Msg("cannot find python3")
		}

		pCmd := exec.Command(python3, tmpfile.Name())
		err = pCmd.Run()
		if err != nil {
			log.Fatal().Err(err).Msg("cannot set default profile")
		}
	},
}

func init() {
	defaultCmd.Flags().StringVarP(&defaultProfileName, "name", "", DefaultProfile, "Name of the profile")

	rootCmd.AddCommand(defaultCmd)
}
