# Slack Integration

Kocho supports Slack notifications. For convenience there is a `kocho slack
init` command. Alternatively just drop a config file in
`~/.giantswarm/kocho/slack.conf`. To use another location provide the
`KOCHO_SLACK_CONFIG_FILE` environment variable. Setting up the configuration
manually, you need to provide the following configuration.  Just replace values
based on your personal needs.

```json
{
    "notification_channel": "channel",
    "token": "token",
    "username": "username"
}
```

You can obtain a slack token using the Slack web user interface.
You need to create a Bot integration.
