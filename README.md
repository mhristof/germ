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
2. Kubernetes from ~/.kube/config`. If there are multiple clusters in the config, it splits out into different files and each profile utilises the extracted config. If you modify ~/.kube/config, you need to re-run this script.


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
