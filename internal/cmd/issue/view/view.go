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
	helpText = `View displays contents of an issue.`
	examples = `$ jira issue view ISSUE-1

# Show 5 recent comments when viewing the issue
$ jira issue view ISSUE-1 --comments 5

# Skip fetching available transitions
$ jira issue view ISSUE-1 --no-transitions

# Get the raw JSON data
$ jira issue view ISSUE-1 --raw`

	flagRaw         = "raw"
	flagDebug       = "debug"
	flagComments    = "comments"
	flagPlain       = "plain"
	flagTransitions = "transitions"

	configProject = "project.key"
	configServer  = "server"

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

	cmd.Flags().Uint(flagComments, 1, "Show N comments")
	cmd.Flags().Bool(flagPlain, false, "Display output in plain mode")
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
	if cmd.Flags().Changed(flagComments) {
		comments, err = cmd.Flags().GetUint(flagComments)
		cmdutil.ExitIfError(err)
	} else {
		numComments := viper.GetUint("num_comments")
		comments = max(numComments, 1)
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
