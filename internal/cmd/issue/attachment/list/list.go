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
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `List prints attachments of an issue with their id, name, size and author.`
	examples = `$ jira issue attachment list ISSUE-1

# Get the data in plain format
$ jira issue attachment list ISSUE-1 --plain

# Get the data in raw json format
$ jira issue attachment list ISSUE-1 --raw`

	flagPlain = "plain"
	flagRaw   = "raw"

	tabWidth   = 8
	tabPadding = 2
)

// NewCmdAttachmentList is an attachment list command.
func NewCmdAttachmentList() *cobra.Command {
	cmd := cobra.Command{
		Use:     "list ISSUE-KEY",
		Short:   "List attachments of an issue",
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

	attachments, err := func() ([]*jira.Attachment, error) {
		s := cmdutil.Info("Fetching attachments...")
		defer s.Stop()

		return api.ProxyIssueAttachments(client, key)
	}()
	cmdutil.ExitIfError(err)

	if raw {
		out, err := json.MarshalIndent(attachments, "", "  ")
		cmdutil.ExitIfError(err)

		fmt.Println(string(out))

		return
	}

	if len(attachments) == 0 {
		cmdutil.Fail("No attachments found for issue %s", key)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, tabWidth, tabPadding, '\t', 0)

	if plain {
		_, _ = fmt.Fprintln(w, "ID\tFILENAME\tSIZE\tAUTHOR")
	} else {
		header := color.New(color.FgGreen, color.Bold).Sprint
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", header("ID"), header("FILENAME"), header("SIZE"), header("AUTHOR"))
	}

	for _, a := range attachments {
		filename := a.Filename
		if !plain {
			filename = color.New(color.FgYellow).Sprint(filename)
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.ID, filename, cmdutil.HumanizeBytes(a.Size), a.Author.DisplayName)
	}

	cmdutil.ExitIfError(w.Flush())
}
