package delete

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue/worklog/pick"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `Delete removes a worklog from an issue. Without a worklog id, the command lets you pick one interactively.`
	examples = `$ jira issue worklog delete ISSUE-1

# Delete a worklog by its id
$ jira issue worklog delete ISSUE-1 10001 --no-input

# Delete a worklog and set a new remaining estimate
$ jira issue worklog delete ISSUE-1 10001 --new-estimate 1d --no-input`
)

// NewCmdWorklogDelete is a worklog delete command.
func NewCmdWorklogDelete() *cobra.Command {
	cmd := cobra.Command{
		Use:     "delete ISSUE-KEY [WORKLOG-ID]",
		Short:   "Delete a worklog of an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"remove", "rm", "del"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1\n" +
				"WORKLOG-ID\tWorklog id, eg: 10001. Omit to pick a worklog interactively",
		},
		Args: cobra.RangeArgs(1, 2),
		Run:  del,
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().String("new-estimate", "", "the new estimate for the backlog to be completed by")
	cmd.Flags().Bool("no-input", false, "Disable prompt for non-required fields")

	return &cmd
}

func del(cmd *cobra.Command, args []string) {
	params := parseArgsAndFlags(args, cmd.Flags())
	client := api.DefaultClient(params.debug)

	worklogs, err := func() ([]*jira.Worklog, error) {
		s := cmdutil.Info("Fetching worklogs...")
		defer s.Stop()

		return client.GetIssueWorklogs(params.issueKey)
	}()
	cmdutil.ExitIfError(err)

	if len(worklogs) == 0 {
		cmdutil.Failed("No worklogs found for issue %q", params.issueKey)
	}
	if params.noInput && params.worklogID == "" {
		cmdutil.Failed("WORKLOG-ID is required with --no-input")
	}

	wl, err := pick.Worklog(worklogs, params.worklogID)
	cmdutil.ExitIfError(err)

	err = func() error {
		s := cmdutil.Info("Removing a worklog")
		defer s.Stop()

		return client.DeleteIssueWorklog(params.issueKey, wl.ID, params.newEstimate)
	}()
	cmdutil.ExitIfError(err)

	cmdutil.Success("Worklog %q removed from issue %q", wl.ID, params.issueKey)
}

type deleteParams struct {
	issueKey    string
	worklogID   string
	newEstimate string
	noInput     bool
	debug       bool
}

func parseArgsAndFlags(args []string, flags query.FlagParser) *deleteParams {
	issueKey := cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])

	var worklogID string
	if len(args) >= 2 {
		worklogID = args[1]
	}

	newEstimate, err := flags.GetString("new-estimate")
	cmdutil.ExitIfError(err)

	noInput, err := flags.GetBool("no-input")
	cmdutil.ExitIfError(err)

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	return &deleteParams{
		issueKey:    issueKey,
		worklogID:   worklogID,
		newEstimate: newEstimate,
		noInput:     noInput,
		debug:       debug,
	}
}
