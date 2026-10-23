package list

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue/worklog/pick"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `List prints worklogs of an issue with their id, start time, time spent, author and comment.`
	examples = `$ jira issue worklog list ISSUE-1

# Get the data in plain format
$ jira issue worklog list ISSUE-1 --plain

# Get the data in raw json format
$ jira issue worklog list ISSUE-1 --raw`

	flagPlain = "plain"
	flagRaw   = "raw"

	tabWidth   = 8
	tabPadding = 2
)

// NewCmdWorklogList is a worklog list command.
func NewCmdWorklogList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list ISSUE-KEY",
		Short:   "List worklogs of an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"ls"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1",
		},
		Args: cobra.ExactArgs(1),
		Run:  list,
	}

	cmd.Flags().Bool(flagPlain, false, "Display output in plain mode")
	cmd.Flags().Bool(flagRaw, false, "Print raw API response instead")

	return &cmd
}

func list(cmd *cobra.Command, args []string) {
	project := viper.GetString("project.key")

	key := cmdutil.GetJiraIssueKey(project, args[0])

	plain, err := cmd.Flags().GetBool(flagPlain)
	cmdutil.ExitIfError(err)

	raw, err := cmd.Flags().GetBool(flagRaw)
	cmdutil.ExitIfError(err)

	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	client := api.DefaultClient(debug)

	worklogs, err := func() ([]*jira.Worklog, error) {
		s := cmdutil.Info("Fetching worklogs...")
		defer s.Stop()

		return client.GetIssueWorklogs(key)
	}()
	cmdutil.ExitIfError(err)

	if raw {
		out, err := json.MarshalIndent(worklogs, "", "  ")
		cmdutil.ExitIfError(err)

		fmt.Println(string(out))

		return
	}

	if len(worklogs) == 0 {
		cmdutil.Fail("No worklogs found for issue %s", key)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, tabWidth, tabPadding, '\t', 0)

	if plain {
		_, _ = fmt.Fprintln(w, "ID\tSTARTED\tTIME SPENT\tAUTHOR\tCOMMENT")
	} else {
		header := color.New(color.FgGreen, color.Bold).Sprint
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", header("ID"), header("STARTED"), header("TIME SPENT"), header("AUTHOR"), header("COMMENT"))
	}

	for _, wl := range worklogs {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", wl.ID, wl.Started, wl.TimeSpent, pick.Author(wl), pick.OneLine(wl.Comment))
	}

	cmdutil.ExitIfError(w.Flush())
}
