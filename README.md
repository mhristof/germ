```
 _______  _______  ______    __   __
|       ||       ||    _ |  |  |_|  |
|    ___||    ___||   | ||  |       |
|   | __ |   |___ |   |_||_ |       |
|   ||  ||    ___||    __  ||       |
|   |_| ||   |___ |   |  | || ||_|| |
|_______||_______||___|  |_||_|   |_|
```

Create iterm2 dynamic profiles.

## Installation

```
go get github.com/mhristof/germ
```

## Coverage

This script extracts profiles for:

1. AWS from `~/.aws/config`
2. Kubernetes from `~/.kube/config`. If there are multiple clusters in the config, it splits out into different files and each profile utilises the extracted config. If you modify `~/.kube/config`, you need to re-run this script.


## Files

germ follows the [XDG base directory spec](https://specifications.freedesktop.org/basedir-spec/latest/),
using `$HOME/.config`, `$HOME/.local/state` and `$HOME/.cache` when the `XDG_*` variables are not set.

| File | Purpose |
|---|---|
| `$XDG_CONFIG_HOME/germ/config.yaml` | your config: custom `profiles`, `ssm.commands`, `aws_open_device_sso` |
| `$XDG_CONFIG_HOME/germ/ssr.json` | your Smart Selection rules |
| `$XDG_STATE_HOME/germ/profiles.json` | generated, keeps the profile badge names stable between runs |
| `$XDG_CACHE_HOME/germ/ssm.json` | generated, SSM profiles used by `generate --ignore-instances` |

Files in the old locations (`~/.config/germ.yaml`, `~/.germ.ssr.json`, `~/.config/germ-profiles.json`
and `~/.cache/germ.ssm.json`) are moved to the new ones the next time germ runs.

The keys under `profiles.<name>.config` are case insensitive, ie `BadgeText` and `badgetext` are the same.

## F.A.Q.

### My custom secret env var is not set.

You need to 'login' after you have opened your custom profile with <kbd>Opt</kbd> + <kbd>a</kbd>.

### How can i store a simple secret in the keychain and use it ?

You can do that via the `new` command, for example

```
germ add --name manos --export
Enter secret:%
```

The export flag will create the secret value to be `export MANOS=%s` where `%s` is the value you typed.

If you want more complicated commands, you can ommit the `--export` flag and type the full text of your secret, for example

```
export FOO=bar BAR=baz
```

### How can I run a command when an SSM session starts ?

Add a list of instance name regexes and commands to `~/.config/germ/config.yaml` (not part of this repo):

```yaml
ssm:
  commands:
    - match: "^marketmaker-apne1-az1-1$"
      command: "cd ~/marketmaker/deltix-mm/; alias d='docker compose'"
    - match: "^marketmaker-"
      command: "cd ~/marketmaker"
```

and re-run `germ generate --write` (`--ignore-instances` is enough, the commands are applied on top of the cached SSM profiles).
`match` is a Go regex matched against the instance `Name` tag; it is unanchored, so use `^...$` for an exact name and `(?i)` for case insensitive matching.
The first matching entry wins and its command is sent once the `Starting session with SessionId:` line is printed.

### My profile doesnt show in the list.

Make sure you have generated `germ generate` and written `--write` your profile and give iterm2
a couple of seconds to detect the changes and use the new profile definitions.

## Custom rules

### SmartSelectionRules

The user can define her own Smart selection rules in a file called `~/.config/germ/ssr.json`. For example

> cat ~/.config/germ/ssr.json
```json
[
    {
      "notes" : "jira ticket link",
      "precision" : "very_high",
      "regex" : "JENKINS-\\d*",
      "actions" : [
        {
          "title" : "Open Jenkins jira link",
          "action" : 1,
          "parameter" : "https://issues.jenkins-ci.org/browse/\\0"
        }
      ]
    },
]
```
