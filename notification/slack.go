// Package notification can be used to send notifications to your Slack team.
package notification

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/juju/errgo"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/nlopes/slack"
)

const (
	defaultConfigPath = "~/.giantswarm/kocho/slack.conf"
)

var (
	configPath string
)

func init() {
	configPath = os.Getenv("KOCHO_SLACK_CONFIG_FILE")
	if configPath == "" {
		configPath = defaultConfigPath
	}
}

// SlackConfiguration describes a configuration for posting to a Slack channel.
type SlackConfiguration struct {
	Token                string `json:"token"`                           // Token is the API token from Slack.
	Username             string `json:"username"`                        // Username is your username.
	NotificationUsername string `json:"notification_username,omitempty"` // NotificationUsername is the username the notification should be posted under.
	EmojiIcon            string `json:"emoji_icon,omitempty"`            // EmojiIcon is an emoji (e.g: :smile:) to use as an avatar for the notification.
	NotificationChannel  string `json:"notification_channel"`            // NotificationChannel is the channel to post to.
}

// WriteConfig writes the given SlackConfiguration to the configuration file.
func WriteConfig(config SlackConfiguration) error {
	expanded, err := homedir.Expand(configPath)
	if err != nil {
		return err
	}

	configFolder := filepath.Dir(expanded)
	if err := os.MkdirAll(configFolder, 0777); err != nil {
		return err
	}
	configFile, err := os.Create(expanded)
	if err != nil {
		return err
	}
	defer configFile.Close()

	if err := json.NewEncoder(configFile).Encode(config); err != nil {
		return err
	}

	return nil
}

// SendMessage reads the configuration file, and posts a message about Kocho's invocation to Slack.
func SendMessage(version, build string) error {
	expanded, err := homedir.Expand(configPath)
	if err != nil {
		return err
	}
	if _, err := os.Stat(expanded); os.IsNotExist(err) {
		return errgo.Mask(ErrNotConfigured, errgo.Any)
	}

	slackConfiguration := SlackConfiguration{
		NotificationUsername: "KochoBot",
		EmojiIcon:            ":robot_face:",
	}

	configFile, err := os.Open(expanded)
	if err != nil {
		return errgo.WithCausef(err, ErrInvalidConfiguration, "couldn't open Slack configuration file")
	}
	defer configFile.Close()

	if err := json.NewDecoder(configFile).Decode(&slackConfiguration); err != nil {
		return errgo.WithCausef(err, ErrInvalidConfiguration, "couldn't decode Slack configuration")
	}

	client := slack.New(slackConfiguration.Token)

	params := slack.PostMessageParameters{}
	params.Attachments = []slack.Attachment{
		slack.Attachment{
			Color: "#2484BE",
			Text:  fmt.Sprintf("*Kocho*: %s ran `%s`", slackConfiguration.Username, strings.Join(os.Args, " ")),
			Fields: []slack.AttachmentField{
				slack.AttachmentField{
					Title: "Kocho Version",
					Value: version,
					Short: true,
				},
				slack.AttachmentField{
					Title: "Kocho Build",
					Value: build,
					Short: true,
				},
			},
			MarkdownIn: []string{"text"},
		},
	}
	params.Username = slackConfiguration.NotificationUsername
	params.IconEmoji = slackConfiguration.EmojiIcon

	if _, _, err := client.PostMessage(slackConfiguration.NotificationChannel, "", params); err != nil {
		return err
	}

	return nil
}
