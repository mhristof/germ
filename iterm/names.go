package iterm

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/mhristof/germ/paths"
	log "github.com/sirupsen/logrus"
)

func newUniqueName(name string) string {
	storage := paths.Names()

	contents, err := os.ReadFile(storage)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"storage": storage,
		}).Debug("Failed to read storage file")
		paths.Write(storage, []byte(fmt.Sprintf(`{"%s": "%s"}`, name, name)))
		return name
	}

	var existing map[string]string
	err = json.Unmarshal(contents, &existing)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"storage": storage,
		}).Debug("Failed to unmarshal storage file")
		return name
	}

	uname, found := existing[name]
	if found {
		return uname
	}

	for {
		uname = fmt.Sprintf("%s-%s", left[rand.Intn(len(left))], right[rand.Intn(len(right))])

		inUse := false
		for _, existingName := range existing {
			if existingName == uname {
				inUse = true
				continue
			}
		}

		if !inUse {
			break
		}
	}

	existing[name] = uname

	prettyJSON, err := json.MarshalIndent(existing, "", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"storage": storage,
		}).Debug("Failed to marshal storage file")
		return name
	}

	err = paths.Write(storage, prettyJSON)
	if err != nil {
		log.WithFields(log.Fields{
			"error":   err,
			"storage": storage,
		}).Error("Failed to write storage file")
	}

	return existing[name]
}
