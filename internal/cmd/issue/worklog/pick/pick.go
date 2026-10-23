package pick

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

// Worklog returns a worklog by its id or lets the user pick one interactively.
func Worklog(worklogs []*jira.Worklog, id string) (*jira.Worklog, error) {
	if id != "" {
		for _, w := range worklogs {
			if w.ID == id {
				return w, nil
			}
		}

		return nil, fmt.Errorf("worklog %q not found in the issue", id)
	}

	var (
		options = make([]string, 0, len(worklogs))
		byEntry = make(map[string]*jira.Worklog, len(worklogs))
	)

	for _, w := range worklogs {
		entry := fmt.Sprintf("%s  %s  %s  %s  %s", w.ID, w.Started, w.TimeSpent, Author(w), OneLine(w.Comment))
		options = append(options, entry)
		byEntry[entry] = w
	}

	var ans string

	if err := survey.AskOne(&survey.Select{Message: "Select worklog", Options: options}, &ans); err != nil {
		return nil, err
	}

	return byEntry[ans], nil
}

// Author returns the worklog author display name.
func Author(w *jira.Worklog) string {
	if w == nil || w.Author == nil {
		return ""
	}
	if w.Author.DisplayName != "" {
		return w.Author.DisplayName
	}
	return w.Author.Name
}

// OneLine replaces line breaks with spaces to keep the output on a single line.
func OneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
