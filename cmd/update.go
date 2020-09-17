package cmd

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"github.com/mhristof/germ/log"
	"github.com/spf13/cobra"
)

var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update the binary with a new version",
		Run: func(cmd *cobra.Command, args []string) {
			url := fmt.Sprintf("https://github.com/mhristof/germ/releases/latest/download/germ.%s", runtime.GOOS)

			latest := wget(url)
			latestSha := sha256.Sum256(latest)

			this, err := os.Executable()
			if err != nil {
				log.WithFields(log.Fields{
					"err": err,
				}).Panic("Cannot retrieve executable")

			}

			f, err := os.Open(this)
			if err != nil {
				log.WithFields(log.Fields{
					"err":  err,
					"this": this,
				}).Panic("Cannot open file")

			}
			defer f.Close()

			thisB, err := ioutil.ReadAll(f)
			thisSha := sha256.Sum256(thisB)

			if thisSha != latestSha {
				fmt.Println("Updating to new version")
				err := ioutil.WriteFile(this, latest, 0755)
				if err != nil {
					log.WithFields(log.Fields{
						"err":  err,
						"this": this,
					}).Panic("Cannot write file")

				}

			}

		},
	}
)

func sha(in []byte) string {
	sum := sha256.Sum256(in)
	return fmt.Sprintf("%x", sum)
}

func wget(url string) []byte {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"url": url,
		}).Panic("Cannot download url")

	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
			"url": url,
		}).Panic("Cannot read response body")

	}

	return data
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
