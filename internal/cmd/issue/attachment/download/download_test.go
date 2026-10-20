package download

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestSelectAttachments(t *testing.T) {
	attachments := []*jira.Attachment{
		{ID: "1", Filename: "a.png"},
		{ID: "2", Filename: "b.pdf"},
	}

	all, err := selectAttachments(attachments, nil, false)
	assert.NoError(t, err)
	assert.Equal(t, attachments, all)

	all, err = selectAttachments(attachments, []string{"a.png"}, true)
	assert.NoError(t, err)
	assert.Equal(t, attachments, all)

	byName, err := selectAttachments(attachments, []string{"b.pdf"}, false)
	assert.NoError(t, err)
	assert.Len(t, byName, 1)
	assert.Equal(t, "2", byName[0].ID)

	byID, err := selectAttachments(attachments, []string{"1"}, false)
	assert.NoError(t, err)
	assert.Len(t, byID, 1)
	assert.Equal(t, "a.png", byID[0].Filename)

	deduped, err := selectAttachments(attachments, []string{"1", "a.png"}, false)
	assert.NoError(t, err)
	assert.Len(t, deduped, 1)

	_, err = selectAttachments(attachments, []string{"missing.png"}, false)
	assert.EqualError(t, err, `attachment "missing.png" not found in the issue`)
}

func TestSafeFilename(t *testing.T) {
	assert.Equal(t, "a.png", safeFilename(&jira.Attachment{ID: "1", Filename: "a.png"}))
	assert.Equal(t, "passwd", safeFilename(&jira.Attachment{ID: "1", Filename: "../../etc/passwd"}))
	assert.Equal(t, "1", safeFilename(&jira.Attachment{ID: "1", Filename: ""}))
}
