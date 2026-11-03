package pick

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
	"github.com/ankitpokhrel/jira-cli/pkg/md"
)

// Comment returns a comment by its id or lets the user pick one interactively.
func Comment(comments []*jira.Comment, id string) (*jira.Comment, error) {
	if id != "" {
		for _, c := range comments {
			if c.ID == id {
				return c, nil
			}
		}

		return nil, fmt.Errorf("comment %q not found in the issue", id)
	}

	var (
		options = make([]string, 0, len(comments))
		byEntry = make(map[string]*jira.Comment, len(comments))
	)

	for _, c := range comments {
		entry := fmt.Sprintf("%s  %s  %s  %s", c.ID, Author(c), c.Created, OneLine(md.FromJiraMD(c.Body)))
		options = append(options, entry)
		byEntry[entry] = c
	}

	var ans string

	if err := survey.AskOne(&survey.Select{Message: "Select comment", Options: options}, &ans); err != nil {
		return nil, err
	}

	return byEntry[ans], nil
}

// Author returns the comment author display name.
func Author(c *jira.Comment) string {
	if c == nil || c.Author == nil {
		return ""
	}
	if c.Author.DisplayName != "" {
		return c.Author.DisplayName
	}
	return c.Author.Name
}

// OneLine replaces line breaks with spaces to keep the output on a single line.
func OneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
