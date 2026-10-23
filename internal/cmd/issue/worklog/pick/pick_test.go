package pick

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestWorklogByID(t *testing.T) {
	worklogs := []*jira.Worklog{
		{ID: "1", TimeSpent: "1h"},
		{ID: "2", TimeSpent: "2h"},
	}

	wl, err := Worklog(worklogs, "2")
	assert.NoError(t, err)
	assert.Equal(t, "2", wl.ID)

	_, err = Worklog(worklogs, "3")
	assert.EqualError(t, err, `worklog "3" not found in the issue`)
}

func TestAuthor(t *testing.T) {
	assert.Equal(t, "", Author(nil))
	assert.Equal(t, "", Author(&jira.Worklog{}))
	assert.Equal(t, "John Doe", Author(&jira.Worklog{Author: &jira.User{DisplayName: "John Doe"}}))
	assert.Equal(t, "jdoe", Author(&jira.Worklog{Author: &jira.User{Name: "jdoe"}}))
}

func TestOneLine(t *testing.T) {
	assert.Equal(t, "first line second line", OneLine("first line\nsecond line"))
	assert.Equal(t, "text", OneLine("  text  "))
}
