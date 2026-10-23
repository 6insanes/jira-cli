package edit

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue/worklog/pick"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
	"github.com/ankitpokhrel/jira-cli/pkg/surveyext"
)

const (
	helpText = `Edit updates a worklog of an issue. Without a worklog id, the command lets you pick one interactively.
Leave a field flag out to keep its current value.`
	examples = `$ jira issue worklog edit ISSUE-1

# Edit a worklog by its id
$ jira issue worklog edit ISSUE-1 10001 --time-spent "1h 30m" --comment "Updated comment" --no-input

# Edit the start date of a worklog
$ jira issue worklog edit ISSUE-1 10001 --started "2022-01-01 09:30:00" --timezone "Europe/Berlin" --no-input

# Edit a worklog and set a new remaining estimate
$ jira issue worklog edit ISSUE-1 10001 --time-spent 2h --new-estimate 1d --no-input`
)

// NewCmdWorklogEdit is a worklog edit command.
func NewCmdWorklogEdit() *cobra.Command {
	cmd := cobra.Command{
		Use:     "edit ISSUE-KEY [WORKLOG-ID]",
		Short:   "Edit a worklog of an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"update"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1\n" +
				"WORKLOG-ID\tWorklog id, eg: 10001. Omit to pick a worklog interactively",
		},
		Args: cobra.RangeArgs(1, 2),
		Run:  edit,
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().String("time-spent", "", "Time to log as days (d), hours (h), or minutes (m), eg: 2d 1h 30m")
	cmd.Flags().String("started", "", "The datetime on which the worklog effort was started, eg: 2022-01-01 09:30:00")
	cmd.Flags().String("timezone", "UTC", "The timezone to use for the started date in IANA timezone format, eg: Europe/Berlin")
	cmd.Flags().String("comment", "", "Comment about the worklog")
	cmd.Flags().String("new-estimate", "", "the new estimate for the backlog to be completed by")
	cmd.Flags().Bool("no-input", false, "Disable prompt for non-required fields")

	return &cmd
}

func edit(cmd *cobra.Command, args []string) {
	params := parseArgsAndFlags(args, cmd)
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

	if !params.noInput && !params.hasChanges() {
		cmdutil.ExitIfError(askQuestions(params, wl))
	}

	if !params.hasChanges() {
		cmdutil.Failed("Nothing to update, use --time-spent, --started or --comment")
	}

	err = func() error {
		s := cmdutil.Info("Updating a worklog")
		defer s.Stop()

		return client.UpdateIssueWorklog(params.issueKey, wl.ID, params.started, params.timeSpent, params.comment, params.newEstimate)
	}()
	cmdutil.ExitIfError(err)

	server := viper.GetString("server")

	cmdutil.Success("Worklog %q updated in issue %q", wl.ID, params.issueKey)
	fmt.Printf("%s\n", cmdutil.GenerateServerBrowseURL(server, params.issueKey))
}

type editParams struct {
	issueKey    string
	worklogID   string
	timeSpent   string
	started     string
	timezone    string
	comment     *string
	newEstimate string
	noInput     bool
	debug       bool
}

func (p *editParams) hasChanges() bool {
	return p.timeSpent != "" || p.started != "" || p.comment != nil
}

func parseArgsAndFlags(args []string, cmd *cobra.Command) *editParams {
	flags := cmd.Flags()

	issueKey := cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])

	var worklogID string
	if len(args) >= 2 {
		worklogID = args[1]
	}

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	timeSpent, err := flags.GetString("time-spent")
	cmdutil.ExitIfError(err)

	started, err := flags.GetString("started")
	cmdutil.ExitIfError(err)

	timezone, err := flags.GetString("timezone")
	cmdutil.ExitIfError(err)

	startedWithTZ, err := cmdutil.DateStringToJiraFormatInLocation(started, timezone)
	cmdutil.ExitIfError(err)

	var comment *string
	if flags.Changed("comment") {
		value, err := flags.GetString("comment")
		cmdutil.ExitIfError(err)
		comment = &value
	}

	newEstimate, err := flags.GetString("new-estimate")
	cmdutil.ExitIfError(err)

	noInput, err := flags.GetBool("no-input")
	cmdutil.ExitIfError(err)

	return &editParams{
		issueKey:    issueKey,
		worklogID:   worklogID,
		timeSpent:   timeSpent,
		started:     startedWithTZ,
		timezone:    timezone,
		comment:     comment,
		newEstimate: newEstimate,
		noInput:     noInput,
		debug:       debug,
	}
}

func askQuestions(params *editParams, wl *jira.Worklog) error {
	qs := []*survey.Question{
		{
			Name: "timeSpent",
			Prompt: &survey.Input{
				Message: "Time spent",
				Default: wl.TimeSpent,
				Help:    "Time to log as days (d), hours (h), or minutes (m), separated by space eg: 2d 1h 30m",
			},
			Validate: survey.Required,
		},
		{
			Name: "started",
			Prompt: &survey.Input{
				Message: "Started",
				Default: wl.Started,
				Help:    "The datetime on which the worklog effort was started, eg: 2022-01-01 09:30:00",
			},
		},
		{
			Name: "comment",
			Prompt: &surveyext.JiraEditor{
				Editor: &survey.Editor{
					Message:       "Comment body",
					Default:       wl.Comment,
					HideDefault:   true,
					AppendDefault: true,
				},
				BlankAllowed: true,
			},
		},
	}

	ans := struct{ TimeSpent, Started, Comment string }{}

	if err := survey.Ask(qs, &ans); err != nil {
		return err
	}

	if ans.TimeSpent != "" && ans.TimeSpent != wl.TimeSpent {
		params.timeSpent = ans.TimeSpent
	}
	if ans.Started != "" && ans.Started != wl.Started {
		started, err := cmdutil.DateStringToJiraFormatInLocation(ans.Started, params.timezone)
		if err != nil {
			return err
		}
		params.started = started
	}
	if ans.Comment != wl.Comment {
		params.comment = &ans.Comment
	}

	return nil
}
