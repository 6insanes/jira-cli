package view

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	tuiView "github.com/ankitpokhrel/jira-cli/internal/view"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
	"github.com/ankitpokhrel/jira-cli/pkg/jira/filter/issue"
)

const (
	helpText = `View displays contents of an issue.

The last 10 comments and the available transitions are printed by default. Use --comments
to change how many comments are shown (0 hides the block) and --transitions=false to skip
the transitions lookup.`
	examples = `$ jira issue view ISSUE-1

# Show 25 recent comments when viewing the issue
$ jira issue view ISSUE-1 --comments 25

# Hide the comments block
$ jira issue view ISSUE-1 --comments 0

# Skip fetching available transitions
$ jira issue view ISSUE-1 --transitions=false

# Get the raw JSON data
$ jira issue view ISSUE-1 --raw`

	flagRaw         = "raw"
	flagDebug       = "debug"
	flagComments    = "comments"
	flagPlain       = "plain"
	flagTransitions = "transitions"

	configProject = "project.key"
	configServer  = "server"

	defaultNumComments = 10

	messageFetchingData = "Fetching issue details..."
)

// NewCmdView is a view command.
func NewCmdView() *cobra.Command {
	cmd := cobra.Command{
		Use:     "view ISSUE-KEY",
		Short:   "View displays contents of an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"show"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1",
		},
		Args: cobra.MinimumNArgs(1),
		Run:  view,
	}

	cmd.Flags().Uint(flagComments, defaultNumComments, "Show N latest comments (0 hides them)")
	cmd.Flags().Bool(flagPlain, false, "Print plain text without colors or pager (suitable for scripts)")
	cmd.Flags().Bool(flagTransitions, true, "Fetch available transitions")
	cmd.Flags().Bool(flagRaw, false, "Print raw Jira API response")

	return &cmd
}

func view(cmd *cobra.Command, args []string) {
	raw, err := cmd.Flags().GetBool(flagRaw)
	cmdutil.ExitIfError(err)

	if raw {
		viewRaw(cmd, args)
		return
	}
	viewPretty(cmd, args)
}

func viewRaw(cmd *cobra.Command, args []string) {
	debug, err := cmd.Flags().GetBool(flagDebug)
	cmdutil.ExitIfError(err)

	key := cmdutil.GetJiraIssueKey(viper.GetString(configProject), args[0])

	apiResp, err := func() (string, error) {
		s := cmdutil.Info(messageFetchingData)
		defer s.Stop()

		client := api.DefaultClient(debug)
		return api.ProxyGetIssueRaw(client, key)
	}()
	cmdutil.ExitIfError(err)

	fmt.Println(apiResp)
}

func viewPretty(cmd *cobra.Command, args []string) {
	debug, err := cmd.Flags().GetBool(flagDebug)
	cmdutil.ExitIfError(err)

	var comments uint
	switch {
	case cmd.Flags().Changed(flagComments):
		comments, err = cmd.Flags().GetUint(flagComments)
		cmdutil.ExitIfError(err)
	case viper.IsSet("num_comments"):
		comments = viper.GetUint("num_comments")
	default:
		comments = defaultNumComments
	}

	key := cmdutil.GetJiraIssueKey(viper.GetString(configProject), args[0])
	iss, trs, err := func() (*jira.Issue, []*jira.Transition, error) {
		s := cmdutil.Info(messageFetchingData)
		defer s.Stop()

		client := api.DefaultClient(debug)

		iss, err := api.ProxyGetIssue(client, key, issue.NewNumCommentsFilter(comments))
		if err != nil {
			return nil, nil, err
		}

		withTransitions, err := cmd.Flags().GetBool(flagTransitions)
		if err != nil {
			return nil, nil, err
		}
		if !withTransitions {
			return iss, nil, nil
		}

		// Transitions are a nice-to-have; don't fail the whole view if they aren't available.
		trs, _ := api.ProxyTransitions(client, key)

		return iss, trs, nil
	}()
	cmdutil.ExitIfError(err)

	plain, err := cmd.Flags().GetBool(flagPlain)
	cmdutil.ExitIfError(err)

	v := tuiView.Issue{
		Server:      viper.GetString(configServer),
		Data:        iss,
		Transitions: trs,
		Display:     tuiView.DisplayFormat{Plain: plain},
		Options:     tuiView.IssueOption{NumComments: comments},
	}
	cmdutil.ExitIfError(v.Render())
}
