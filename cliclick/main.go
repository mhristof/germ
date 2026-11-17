package cliclick

import (
	"fmt"
	"os/exec"

	"github.com/mhristof/germ/iterm"
	"github.com/pkg/errors"
)

// Profile returns a profile that nudges the mouse to its current position
// every 10 seconds, which keeps the computer from going to sleep.
func Profile() (iterm.Profile, error) {
	path, err := exec.LookPath("cliclick")
	if err != nil {
		return iterm.Profile{}, errors.Wrap(err, "cannot find cliclick binary")
	}

	p := iterm.NewProfile("cliclick", map[string]string{
		"Command": fmt.Sprintf(`/usr/bin/env CLICLICK=%s /bin/sh -c 'echo "keeping the computer awake, ctrl+c to stop"; while sleep 10; do pos=$("$CLICLICK" p); "$CLICLICK" m:$pos; echo "$(date +%%Y-%%m-%%dT%%H:%%M:%%S) nudged mouse at $pos"; done'`, path),
	})

	p.KeyboardMap = map[string]iterm.KeyboardMap{
		"0x77-0x100000-0xd": {
			Version: 1,
			Action:  12,
			Text:    "Cmd+w is disabled, please ctrl+c to exit\n",
		},
	}

	return *p, nil
}
