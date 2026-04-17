package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mhristof/germ/aws"
	"github.com/mhristof/germ/iterm"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type alfredItem struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Arg      string `json:"arg"`
	Match    string `json:"match"`
}

type alfredOutput struct {
	Items []alfredItem `json:"items"`
}

var alfredCmd = &cobra.Command{
	Use:   "alfred",
	Short: "Generate Alfred Script Filter JSON for AWS profiles",
	Run: func(cmd *cobra.Command, args []string) {
		profiles := aws.Profiles("", AWSConfig)

		items := make([]alfredItem, 0, len(profiles))
		for _, prof := range profiles {
			if strings.HasPrefix(prof.Name, "login-") {
				continue
			}

			items = append(items, alfredItem{
				Title:    prof.Name,
				Subtitle: buildSubtitle(prof),
				Arg:      prof.Name,
				Match:    buildMatch(prof),
			})
		}

		output := alfredOutput{Items: items}

		data, err := json.Marshal(output)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot marshal alfred output")
		}

		fmt.Print(string(data))
	},
}

func buildSubtitle(prof iterm.Profile) string {
	var parts []string

	for _, tag := range prof.Tags {
		if strings.HasPrefix(tag, "account=") {
			parts = append(parts, tag)
			continue
		}

		if strings.HasPrefix(tag, "source-profile=") {
			parts = append(parts, tag)
			continue
		}
	}

	region, _ := prof.FindTag("region")
	if region != "" {
		parts = append(parts, region)
	}

	if len(parts) == 0 {
		return "AWS profile"
	}

	return strings.Join(parts, " ")
}

func buildMatch(prof iterm.Profile) string {
	parts := []string{prof.Name}
	parts = append(parts, prof.Tags...)
	return strings.Join(parts, " ")
}

func init() {
	rootCmd.AddCommand(alfredCmd)
}
