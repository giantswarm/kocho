package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/giantswarm/kocho/notification"
)

// Constants related to Slack.
const (
	LabelToken                = "Slack token"
	LabelUsername             = "Your username"
	LabelNotificationUsername = "Bot username (optional)"
	LabelEmoji                = "Bot emoji (optional)"
	LabelNotificationChannel  = "Notification channel"
)

var (
	cmdSlack = &Command{
		Name:        "slack",
		Description: "Set up and test Slack notifications",
		Summary:     "Set up Slack configuration and test that notifications are working",
		Usage:       "init|test",
		Run:         runSlack,
	}
)

func readFromStdin(label string) string {
	fmt.Printf("%s: ", label)

	stdInReader := bufio.NewReader(os.Stdin)
	line, _, _ := stdInReader.ReadLine()

	return string(line)
}

func runSlack(args []string) (exit int) {
	if len(args) != 1 {
		return exitError("usage: kocho slack init|test")
	}

	switch args[0] {
	case "init":
		slackConfiguration := notification.SlackConfiguration{}

		slackConfiguration.Token = readFromStdin(LabelToken)
		slackConfiguration.Username = readFromStdin(LabelUsername)
		slackConfiguration.NotificationUsername = readFromStdin(LabelNotificationUsername)
		slackConfiguration.EmojiIcon = readFromStdin(LabelEmoji)
		slackConfiguration.NotificationChannel = readFromStdin(LabelNotificationChannel)

		if !strings.HasPrefix(slackConfiguration.NotificationChannel, "#") {
			slackConfiguration.NotificationChannel = "#" + slackConfiguration.NotificationChannel
		}

		if err := notification.WriteConfig(slackConfiguration); err != nil {
			return exitError("failed to write configuration: ", err)
		}
	case "test":
		if err := notification.SendMessage(projectVersion, projectBuild); err != nil {
			return exitError("failed to send test message")
		}

		fmt.Println("test message sent successfully")
	}

	return 0
}
