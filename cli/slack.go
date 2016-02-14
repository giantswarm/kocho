package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

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
	slackCmd = &cobra.Command{
		Use:   "slack",
		Short: "Manage Slack notifications",
		Long:  "Manage set up and testing of Slack notifications",
	}

	slackInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialise Slack notifications configuration",
		Long:  "Interactively set up the configuration file for Slack notifications",
		Run:   runSlackInit,
	}

	slackTestCmd = &cobra.Command{
		Use:   "test",
		Short: "Test Slack notifications",
		Long:  "Test Slack notifications by firing a notification",
		Run:   runSlackTest,
	}
)

func init() {
	RootCmd.AddCommand(slackCmd)

	slackCmd.AddCommand(slackInitCmd)
	slackCmd.AddCommand(slackTestCmd)
}

func readFromStdin(label string) string {
	fmt.Printf("%s: ", label)

	stdInReader := bufio.NewReader(os.Stdin)
	line, _, _ := stdInReader.ReadLine()

	return string(line)
}

func runSlackInit(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		cmd.Usage()
		return
	}

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
		fmt.Printf("failed to write configuration: %s\n", err)
		return
	}
}

func runSlackTest(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		cmd.Usage()
		return
	}

	if err := notification.SendMessage(projectVersion, projectBuild); err != nil {
		if notification.IsNotConfigured(err) {
			fmt.Println("Notifications not configured. Use 'kocho slack init'")
			return
		} else if notification.IsInvalidConfiguration(err) {
			fmt.Printf("Invalid configuration file: %s\n", err)
			return
		} else {
			fmt.Printf("Failed to send message: %s\n", err)
			return
		}
	} else {
		fmt.Println("test message sent successfully.")
	}
}

func fireNotification() {
	if err := notification.SendMessage(projectVersion, projectBuild); err != nil {
		if notification.IsNotConfigured(err) {
			fmt.Println("Notifications not configured. Use 'kocho slack init'")
			return
		} else if notification.IsInvalidConfiguration(err) {
			fmt.Printf("Invalid configuration file: %s\n", err)
			return
		} else {
			fmt.Printf("failed to send message: %s\n", err)
			return
		}
	}
}
