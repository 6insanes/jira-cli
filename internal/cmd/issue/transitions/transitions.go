package transitions

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `Transitions lists available transitions for an issue with their target states.`
	examples = `$ jira issue transitions ISSUE-1

# Get the data in plain format
$ jira issue transitions ISSUE-1 --plain`

	flagPlain = "plain"
)

// NewCmdTransitions is a transitions command.
func NewCmdTransitions() *cobra.Command {
	cmd := cobra.Command{
		Use:     "transitions ISSUE-KEY",
		Short:   "List available transitions for an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"trs"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1",
		},
		Args: cobra.MinimumNArgs(1),
		Run:  transitions,
	}

	cmd.Flags().Bool(flagPlain, false, "Display output in plain mode")

	return &cmd
}

func transitions(cmd *cobra.Command, args []string) {
	project := viper.GetString("project.key")

	key := cmdutil.GetJiraIssueKey(project, args[0])

	plain, err := cmd.Flags().GetBool(flagPlain)
	cmdutil.ExitIfError(err)

	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	client := api.DefaultClient(debug)

	trs, err := func() ([]*jira.Transition, error) {
		s := cmdutil.Info("Fetching available transitions...")
		defer s.Stop()

		return api.ProxyTransitions(client, key)
	}()
	cmdutil.ExitIfError(err)

	if len(trs) == 0 {
		cmdutil.Fail("No transitions available for issue %s", key)
		return
	}

	maxNameLen := 0
	for _, t := range trs {
		maxNameLen = max(len(t.Name), maxNameLen)
	}

	for _, t := range trs {
		name := t.Name
		if !plain {
			name = color.New(color.FgGreen, color.Bold).Sprint(name)
		}
		target := t.To.Name
		if target == "" {
			target = "(unknown)"
		}
		fmt.Printf("%s%s -> %s\n", name, strings.Repeat(" ", maxNameLen-len(t.Name)), target)
	}
}
