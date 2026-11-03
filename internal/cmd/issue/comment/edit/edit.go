package edit

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue/comment/pick"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
	"github.com/ankitpokhrel/jira-cli/pkg/md"
	"github.com/ankitpokhrel/jira-cli/pkg/surveyext"
)

const (
	helpText = `Edit updates a comment of an issue. Without a comment id, the command lets you pick one interactively.`
	examples = `$ jira issue comment edit ISSUE-1

# Edit a comment by its id
$ jira issue comment edit ISSUE-1 10001 --body "Updated comment" --no-input

# Multi-line comment
$ jira issue comment edit ISSUE-1 10001 --body $'Supports\n\nNew line' --no-input`
)

// NewCmdCommentEdit is a comment edit command.
func NewCmdCommentEdit() *cobra.Command {
	cmd := cobra.Command{
		Use:     "edit ISSUE-KEY [COMMENT-ID]",
		Short:   "Edit a comment of an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"update"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1\n" +
				"COMMENT-ID\tComment id, eg: 10001. Omit to pick a comment interactively",
		},
		Args: cobra.RangeArgs(1, 2),
		Run:  edit,
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().String("body", "", "New body of the comment")
	cmd.Flags().Bool("no-input", false, "Disable prompt for non-required fields")

	return &cmd
}

func edit(cmd *cobra.Command, args []string) {
	params := parseArgsAndFlags(args, cmd.Flags())
	client := api.DefaultClient(params.debug)

	comments, err := func() ([]*jira.Comment, error) {
		s := cmdutil.Info("Fetching comments...")
		defer s.Stop()

		return client.GetIssueComments(params.issueKey)
	}()
	cmdutil.ExitIfError(err)

	if len(comments) == 0 {
		cmdutil.Failed("No comments found for issue %q", params.issueKey)
	}
	if params.noInput && params.commentID == "" {
		cmdutil.Failed("COMMENT-ID is required with --no-input")
	}
	if params.noInput && params.body == "" {
		cmdutil.Failed("--body is required with --no-input")
	}

	comment, err := pick.Comment(comments, params.commentID)
	cmdutil.ExitIfError(err)

	if params.body == "" {
		cmdutil.ExitIfError(askBody(params, comment))
	}

	if strings.TrimSpace(params.body) == "" {
		cmdutil.Failed("Comment body cannot be empty")
	}
	if params.body == md.FromJiraMD(comment.Body) {
		cmdutil.Failed("Nothing to update")
	}

	err = func() error {
		s := cmdutil.Info("Updating a comment")
		defer s.Stop()

		return client.UpdateIssueComment(params.issueKey, comment.ID, params.body)
	}()
	cmdutil.ExitIfError(err)

	server := viper.GetString("server")

	cmdutil.Success("Comment %q updated in issue %q", comment.ID, params.issueKey)
	fmt.Printf("%s\n", cmdutil.GenerateServerBrowseURL(server, params.issueKey))
}

type editParams struct {
	issueKey  string
	commentID string
	body      string
	noInput   bool
	debug     bool
}

func parseArgsAndFlags(args []string, flags query.FlagParser) *editParams {
	issueKey := cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])

	var commentID string
	if len(args) >= 2 {
		commentID = args[1]
	}

	body, err := flags.GetString("body")
	cmdutil.ExitIfError(err)

	noInput, err := flags.GetBool("no-input")
	cmdutil.ExitIfError(err)

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	return &editParams{
		issueKey:  issueKey,
		commentID: commentID,
		body:      body,
		noInput:   noInput,
		debug:     debug,
	}
}

func askBody(params *editParams, comment *jira.Comment) error {
	return survey.AskOne(&surveyext.JiraEditor{
		Editor: &survey.Editor{
			Message:       "Comment body",
			Default:       md.FromJiraMD(comment.Body),
			HideDefault:   true,
			AppendDefault: true,
		},
		BlankAllowed: false,
	}, &params.body)
}
