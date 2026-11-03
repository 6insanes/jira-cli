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
	"github.com/ankitpokhrel/jira-cli/internal/cmd/issue/comment/pick"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
	"github.com/ankitpokhrel/jira-cli/pkg/md"
)

const (
	helpText = `List prints comments of an issue with their id, author, creation date and body.`
	examples = `$ jira issue comment list ISSUE-1

# Get the data in plain format
$ jira issue comment list ISSUE-1 --plain

# Get the data in raw json format
$ jira issue comment list ISSUE-1 --raw`

	flagPlain = "plain"
	flagRaw   = "raw"

	tabWidth   = 8
	tabPadding = 2
)

// NewCmdCommentList is a comment list command.
func NewCmdCommentList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list ISSUE-KEY",
		Short:   "List comments of an issue",
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

	comments, err := func() ([]*jira.Comment, error) {
		s := cmdutil.Info("Fetching comments...")
		defer s.Stop()

		return client.GetIssueComments(key)
	}()
	cmdutil.ExitIfError(err)

	if raw {
		out, err := json.MarshalIndent(comments, "", "  ")
		cmdutil.ExitIfError(err)

		fmt.Println(string(out))

		return
	}

	if len(comments) == 0 {
		cmdutil.Fail("No comments found for issue %s", key)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, tabWidth, tabPadding, '\t', 0)

	if plain {
		_, _ = fmt.Fprintln(w, "ID\tAUTHOR\tCREATED\tBODY")
	} else {
		header := color.New(color.FgGreen, color.Bold).Sprint
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", header("ID"), header("AUTHOR"), header("CREATED"), header("BODY"))
	}

	for _, c := range comments {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", c.ID, pick.Author(c), c.Created, pick.OneLine(md.FromJiraMD(c.Body)))
	}

	cmdutil.ExitIfError(w.Flush())
}
