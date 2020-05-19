![germ](noun_Germ_2997094.png)

Create iterm2 dynamic profiles.

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
## Icon

[Germ by Binpodo from the Noun Project](https://thenounproject.com/search/?q=germ&i=2997094)
