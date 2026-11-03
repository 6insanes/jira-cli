package delete

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue/comment/pick"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `Delete removes a comment from an issue. Without a comment id, the command lets you pick one interactively.`
	examples = `$ jira issue comment delete ISSUE-1

# Delete a comment by its id
$ jira issue comment delete ISSUE-1 10001 --no-input`
)

// NewCmdCommentDelete is a comment delete command.
func NewCmdCommentDelete() *cobra.Command {
	cmd := cobra.Command{
		Use:     "delete ISSUE-KEY [COMMENT-ID]",
		Short:   "Delete a comment of an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"remove", "rm", "del"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1\n" +
				"COMMENT-ID\tComment id, eg: 10001. Omit to pick a comment interactively",
		},
		Args: cobra.RangeArgs(1, 2),
		Run:  del,
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().Bool("no-input", false, "Disable prompt for non-required fields")

	return &cmd
}

func del(cmd *cobra.Command, args []string) {
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

	comment, err := pick.Comment(comments, params.commentID)
	cmdutil.ExitIfError(err)

	if !params.noInput {
		var confirm bool

		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Delete comment %q?", comment.ID),
			Default: false,
		}
		cmdutil.ExitIfError(survey.AskOne(prompt, &confirm))

		if !confirm {
			cmdutil.Failed("Action aborted")
		}
	}

	err = func() error {
		s := cmdutil.Info("Removing a comment")
		defer s.Stop()

		return client.DeleteIssueComment(params.issueKey, comment.ID)
	}()
	cmdutil.ExitIfError(err)

	cmdutil.Success("Comment %q removed from issue %q", comment.ID, params.issueKey)
}

type deleteParams struct {
	issueKey  string
	commentID string
	noInput   bool
	debug     bool
}

func parseArgsAndFlags(args []string, flags query.FlagParser) *deleteParams {
	issueKey := cmdutil.GetJiraIssueKey(viper.GetString("project.key"), args[0])

	var commentID string
	if len(args) >= 2 {
		commentID = args[1]
	}

	noInput, err := flags.GetBool("no-input")
	cmdutil.ExitIfError(err)

	debug, err := flags.GetBool("debug")
	cmdutil.ExitIfError(err)

	return &deleteParams{
		issueKey:  issueKey,
		commentID: commentID,
		noInput:   noInput,
		debug:     debug,
	}
}
