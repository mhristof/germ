package vim

import "github.com/mhristof/germ/iterm"

func Profile() iterm.Profile {
	p := iterm.NewProfile("vim", map[string]string{})

	p.Triggers = nil
	p.BoundHosts = []string{
		"&vim",
	}

	return *p
}
