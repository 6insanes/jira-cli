package jira

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetIssueWorklogs(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/rest/api/2/issue/TEST-1/worklog", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		if unexpectedStatusCode {
			w.WriteHeader(400)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"startAt":0,"maxResults":1048576,"total":1,"worklogs":[{"id":"10001","issueId":"10000","author":{"name":"jdoe","displayName":"John Doe"},"comment":"comment","started":"2022-01-01T01:02:02.000+0200","timeSpent":"1h","timeSpentSeconds":3600}]}`))
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	worklogs, err := client.GetIssueWorklogs("TEST-1")
	assert.NoError(t, err)
	assert.Len(t, worklogs, 1)
	assert.Equal(t, "10001", worklogs[0].ID)
	assert.Equal(t, "John Doe", worklogs[0].Author.DisplayName)
	assert.Equal(t, "1h", worklogs[0].TimeSpent)
	assert.Equal(t, 3600, worklogs[0].TimeSpentSeconds)

	unexpectedStatusCode = true

	_, err = client.GetIssueWorklogs("TEST-1")
	assert.Error(t, &ErrUnexpectedResponse{}, err)
}

func TestUpdateIssueWorklog(t *testing.T) {
	var (
		unexpectedStatusCode bool
		bodies               []string
		queries              []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/rest/api/2/issue/TEST-1/worklog/10001", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		bodies = append(bodies, string(body))
		queries = append(queries, r.URL.RawQuery)

		if unexpectedStatusCode {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	comment := "comment"

	err := client.UpdateIssueWorklog("TEST-1", "10001", "2022-01-01T01:02:02.000+0200", "1h", &comment, "")
	assert.NoError(t, err)

	empty := ""

	err = client.UpdateIssueWorklog("TEST-1", "10001", "", "", &empty, "1d")
	assert.NoError(t, err)

	assert.Equal(t, []string{
		`{"started":"2022-01-01T01:02:02.000+0200","timeSpent":"1h","comment":"comment"}`,
		`{"comment":""}`,
	}, bodies)
	assert.Equal(t, []string{"", "adjustEstimate=new&newEstimate=1d"}, queries)

	unexpectedStatusCode = true

	err = client.UpdateIssueWorklog("TEST-1", "10001", "2022-01-01T01:02:02.000+0200", "1h", nil, "")
	assert.Error(t, &ErrUnexpectedResponse{}, err)
}

func TestDeleteIssueWorklog(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/rest/api/2/issue/TEST-1/worklog/10001", r.URL.Path)

		expectedQuery := ""
		if r.URL.RawQuery != "" {
			expectedQuery = "adjustEstimate=new&newEstimate=1d"
		}
		assert.Equal(t, expectedQuery, r.URL.RawQuery)

		if unexpectedStatusCode {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	err := client.DeleteIssueWorklog("TEST-1", "10001", "")
	assert.NoError(t, err)

	err = client.DeleteIssueWorklog("TEST-1", "10001", "1d")
	assert.NoError(t, err)

	unexpectedStatusCode = true

	err = client.DeleteIssueWorklog("TEST-1", "10001", "")
	assert.Error(t, &ErrUnexpectedResponse{}, err)
}

func TestWorklogEstimateQuery(t *testing.T) {
	assert.Equal(t, "adjustEstimate=new&newEstimate=1d+2h", worklogEstimateQuery("1d 2h"))
	assert.False(t, strings.Contains(worklogEstimateQuery("1d 2h"), " "))
}
