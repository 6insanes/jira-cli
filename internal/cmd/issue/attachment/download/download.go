package download

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
	helpText = `Download saves attachments of an issue into a local directory. Without explicit
attachment names or ids all attachments of the issue are downloaded.`

	examples = `$ jira issue attachment download ISSUE-1

# Download a single attachment by name or id
$ jira issue attachment download ISSUE-1 screenshot.png

# Download all attachments into a specific directory
$ jira issue attachment download ISSUE-1 --all -d ./attachments

# Overwrite files that already exist
$ jira issue attachment download ISSUE-1 10700 --overwrite`

	flagDir       = "dir"
	flagAll       = "all"
	flagOverwrite = "overwrite"

	dirPerm = 0o755
)

// NewCmdAttachmentDownload is an attachment download command.
func NewCmdAttachmentDownload() *cobra.Command {
	cmd := cobra.Command{
		Use:     "download ISSUE-KEY [ATTACHMENT...]",
		Short:   "Download attachments of an issue",
		Long:    helpText,
		Example: examples,
		Aliases: []string{"dl"},
		Annotations: map[string]string{
			"help:args": "ISSUE-KEY\tIssue key, eg: ISSUE-1\n\nATTACHMENT\tAttachment id or filename, eg: 10700 or screenshot.png",
		},
		Args: cobra.MinimumNArgs(1),
		Run:  download,
	}

	cmd.Flags().StringP(flagDir, "d", ".", "Directory to save the attachments into")
	cmd.Flags().Bool(flagAll, false, "Download all attachments of the issue")
	cmd.Flags().Bool(flagOverwrite, false, "Overwrite files that already exist")

	return &cmd
}

func download(cmd *cobra.Command, args []string) {
	project := viper.GetString("project.key")

	key := cmdutil.GetJiraIssueKey(project, args[0])

	dir, err := cmd.Flags().GetString(flagDir)
	cmdutil.ExitIfError(err)

	all, err := cmd.Flags().GetBool(flagAll)
	cmdutil.ExitIfError(err)

	overwrite, err := cmd.Flags().GetBool(flagOverwrite)
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

	if len(attachments) == 0 {
		cmdutil.Fail("No attachments found for issue %s", key)
		return
	}

	selected, err := selectAttachments(attachments, args[1:], all)
	cmdutil.ExitIfError(err)

	cmdutil.ExitIfError(os.MkdirAll(dir, dirPerm))

	for _, a := range selected {
		path := filepath.Join(dir, safeFilename(a))

		if !overwrite {
			if _, statErr := os.Stat(path); statErr == nil {
				cmdutil.Warn("Skipping %s, file already exists (use --overwrite to replace it)", path)
				continue
			}
		}

		size, err := save(client, a, path)
		if err != nil {
			cmdutil.Failed("Failed to download %s: %s", a.Filename, cmdutil.NormalizeJiraError(err.Error()))
		}

		cmdutil.Success("Downloaded %s (%s)", path, cmdutil.HumanizeBytes(size))
	}
}

func selectAttachments(attachments []*jira.Attachment, names []string, all bool) ([]*jira.Attachment, error) {
	if all || len(names) == 0 {
		return attachments, nil
	}

	selected := make([]*jira.Attachment, 0, len(names))
	seen := make(map[string]bool, len(names))

	for _, name := range names {
		var found *jira.Attachment

		for _, a := range attachments {
			if a.ID == name || a.Filename == name {
				found = a
				break
			}
		}

		if found == nil {
			return nil, fmt.Errorf("attachment %q not found in the issue", name)
		}
		if seen[found.ID] {
			continue
		}

		seen[found.ID] = true
		selected = append(selected, found)
	}

	return selected, nil
}

func save(client *jira.Client, a *jira.Attachment, path string) (int64, error) {
	s := cmdutil.Info(fmt.Sprintf("Downloading %s...", a.Filename))
	defer s.Stop()

	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}

	if err = client.DownloadAttachment(a, f); err != nil {
		_ = f.Close()
		_ = os.Remove(path)

		return 0, err
	}

	if err = f.Close(); err != nil {
		_ = os.Remove(path)

		return 0, err
	}

	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

func safeFilename(a *jira.Attachment) string {
	name := filepath.Base(filepath.Clean(string(filepath.Separator) + a.Filename))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return a.ID
	}

	return name
}
