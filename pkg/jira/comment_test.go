package jira

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//nolint:dupl // test scaffolding mirrors the worklog test on purpose.
func TestGetIssueComments(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/rest/api/2/issue/TEST-1/comment", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		if unexpectedStatusCode {
			w.WriteHeader(400)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"startAt":0,"maxResults":1000,"total":2,"comments":[{"id":"10001","author":{"name":"jdoe","displayName":"John Doe"},"body":"comment","created":"2022-01-01T01:02:02.000+0200","updated":"2022-01-01T01:02:02.000+0200"},{"id":"10002","author":{"name":"jdoe","displayName":"John Doe"},"body":"another","created":"2022-01-02T01:02:02.000+0200","updated":"2022-01-02T01:02:02.000+0200"}]}`))
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	comments, err := client.GetIssueComments("TEST-1")
	assert.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.Equal(t, "10001", comments[0].ID)
	assert.Equal(t, "John Doe", comments[0].Author.DisplayName)
	assert.Equal(t, "comment", comments[0].Body)
	assert.Equal(t, "10002", comments[1].ID)

	unexpectedStatusCode = true

	_, err = client.GetIssueComments("TEST-1")
	assert.Error(t, &ErrUnexpectedResponse{}, err)
}

func TestGetIssueCommentsPaginated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/rest/api/2/issue/TEST-1/comment", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)

		if r.URL.Query().Get("startAt") == "0" {
			_, _ = w.Write([]byte(`{"startAt":0,"maxResults":1,"total":2,"comments":[{"id":"10001","body":"first"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"startAt":1,"maxResults":1,"total":2,"comments":[{"id":"10002","body":"second"}]}`))
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	comments, err := client.GetIssueComments("TEST-1")
	assert.NoError(t, err)
	assert.Len(t, comments, 2)
	assert.Equal(t, "10001", comments[0].ID)
	assert.Equal(t, "10002", comments[1].ID)
}

func TestUpdateIssueComment(t *testing.T) {
	var (
		unexpectedStatusCode bool
		body                 string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/rest/api/2/issue/TEST-1/comment/10001", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		actualBody, err := io.ReadAll(r.Body)
		assert.NoError(t, err)
		body = string(actualBody)

		if unexpectedStatusCode {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	err := client.UpdateIssueComment("TEST-1", "10001", "comment")
	assert.NoError(t, err)
	assert.Equal(t, `{"body":"comment"}`, body)

	unexpectedStatusCode = true

	err = client.UpdateIssueComment("TEST-1", "10001", "comment")
	assert.Error(t, &ErrUnexpectedResponse{}, err)
}

func TestDeleteIssueComment(t *testing.T) {
	var unexpectedStatusCode bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/rest/api/2/issue/TEST-1/comment/10001", r.URL.Path)

		if unexpectedStatusCode {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	err := client.DeleteIssueComment("TEST-1", "10001")
	assert.NoError(t, err)

	unexpectedStatusCode = true

	err = client.DeleteIssueComment("TEST-1", "10001")
	assert.Error(t, &ErrUnexpectedResponse{}, err)
}
