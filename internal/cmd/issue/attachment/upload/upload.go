package upload

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `Upload attaches local files to an issue. Every positional argument after
the issue key is treated as a path to a file on your machine.`

	examples = `$ jira issue attachment upload ISSUE-1 screenshot.png

# Upload multiple files at once
$ jira issue attachment upload ISSUE-1 screenshot.png report.pdf

# Upload a file from another directory
$ jira issue attachment upload ISSUE-1 ./logs/application.log`
)

// NewCmdAttachmentUpload is an attachment upload command.
func NewCmdAttachmentUpload() *cobra.Command {
	cmd := cobra.Command{
		Use:     "upload ISSUE-KEY FILE...",
		Short:   "Upload files to an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"add", "up"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1\n\nFILE\tPath to a local file, eg: ./screenshot.png",
		},
		Args: cobra.MinimumNArgs(2),
		Run:  upload,
	}

	return &cmd
}

func upload(cmd *cobra.Command, args []string) {
	project := viper.GetString("project.key")

	key := cmdutil.GetJiraIssueKey(project, args[0])

	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	client := api.DefaultClient(debug)

	var failed int

	for _, path := range args[1:] {
		if err := validateFile(path); err != nil {
			cmdutil.Fail("%s", err)

			failed++

			continue
		}

		attachment, err := func() (*jira.Attachment, error) {
			s := cmdutil.Info(fmt.Sprintf("Uploading %s...", filepath.Base(path)))
			defer s.Stop()

			return api.ProxyUploadAttachment(client, key, path)
		}()
		if err != nil {
			cmdutil.Fail("Failed to upload %s: %s", path, cmdutil.NormalizeJiraError(err.Error()))

			failed++

			continue
		}

		cmdutil.Success("Uploaded %s (%s, id %s)", attachment.Filename, cmdutil.HumanizeBytes(attachment.Size), attachment.ID)
	}

	if failed > 0 {
		os.Exit(1)
	}
}

func validateFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", path)
	}

	return nil
}
