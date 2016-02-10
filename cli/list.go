package cli

import (
	"fmt"
	"time"

	"github.com/ryanuber/columnize"
)

var cmdList = &Command{
	Name:        "list",
	Description: "List all swarms",
	Summary:     "List all existing swarms.",
	Run:         runList,
}

const (
	swarmListHeader = "Name | Type | Created"
	swarmListScheme = "%s | %s | %s"
)

func runList(args []string) (exit int) {
	if len(args) > 0 {
		return exitError("too many arguments")
	}

	swarms, err := swarmService.List()
	if err != nil {
		return exitError("couldn't list swarms", err)
	}
	lines := []string{swarmListHeader}
	for _, s := range swarms {
		lines = append(lines, fmt.Sprintf(swarmListScheme, s.Name, s.Type, s.Created.Format(time.RFC822)))
	}
	fmt.Println(columnize.SimpleFormat(lines))
	return 0
}
