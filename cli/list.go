package cli

import (
	"fmt"
	"time"

	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

const (
	swarmListHeader = "Name | Type | Created"
	swarmListScheme = "%s | %s | %s"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all swarms",
	Long:  "Print out a list of all swarms, with their type and creation time",
	Run:   runList,
}

func init() {
	RootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	if len(args) > 0 {
		cmd.Usage()
		return
	}

	swarms, err := swarmService.List()
	if err != nil {
		fmt.Printf("couldn't list swarms: %s\n", err)
		return
	}

	lines := []string{swarmListHeader}
	for _, s := range swarms {
		lines = append(lines, fmt.Sprintf(swarmListScheme, s.Name, s.Type, s.Created.Format(time.RFC822)))
	}

	fmt.Println(columnize.SimpleFormat(lines))
}
