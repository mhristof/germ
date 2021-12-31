package vault

import (
	"fmt"
	"os/exec"

	"github.com/mhristof/germ/iterm"
	"github.com/pkg/errors"
)

func Profile() (iterm.Profile, error) {
	path, err := exec.LookPath("vault")
	if err != nil {
		return iterm.Profile{}, errors.Wrap(err, "cannot find vault binary")
	}

	p := iterm.NewProfile("vault", map[string]string{
		"Command": fmt.Sprintf("%s server -dev", path),
	})

	p.Triggers = []iterm.Trigger{
		{
			Action:    "CoprocessTrigger",
			Parameter: `echo '\1' > ~/.vault-token`,
			Regex:     "^Root Token: (s.*)",
			Partial:   true,
		},
	}

	p.BoundHosts = []string{
		"&vault",
	}

	p.KeyboardMap = map[string]iterm.KeyboardMap{}

	return *p, nil
}
