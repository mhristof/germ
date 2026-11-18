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

Add the instance name and the command to `~/.config/germ.yaml` (not part of this repo):

```yaml
ssm:
  commands:
    marketmaker-apne1-az1-1: "cd ~/marketmaker/deltix-mm/; alias d='docker compose'"
```

and re-run `germ generate --write` (`--ignore-instances` is enough, the commands are applied on top of the cached SSM profiles).
The command is sent once the `Starting session with SessionId:` line is printed. Instance names are matched case insensitively and must not contain `.`.

### My profile doesnt show in the list.

Make sure you have generated `germ generate` and written `--write` your profile and give iterm2
a couple of seconds to detect the changes and use the new profile definitions.

## Custom rules

### SmartSelectionRules

The user can define her own Smart selection rules in a file called `~/.germ.ssr.json`. For example

> cat ~/.germ.ssr.json
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
