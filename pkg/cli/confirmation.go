package cli

import (
	"eager/pkg/jira/model"
	"fmt"
)

func Confirmation(worklog model.Worklog) bool {
	var answer string

	fmt.Printf("Remove %s (y/N): ", worklog.String())
	_, err := fmt.Scanln(&answer)
	if err != nil {
		return false
	}
	if answer == "y" {
		return true
	}
	return false
}
