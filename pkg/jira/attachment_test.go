package jira

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAttachments(t *testing.T) {
	var (
		apiVersion2          bool
		unexpectedStatusCode bool
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiVersion2 {
			assert.Equal(t, "/rest/api/2/issue/TEST", r.URL.Path)
		} else {
			assert.Equal(t, "/rest/api/3/issue/TEST", r.URL.Path)
		}

		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "attachment", r.URL.Query().Get("fields"))

		if unexpectedStatusCode {
			w.WriteHeader(400)
		} else {
			resp, err := os.ReadFile("./testdata/attachments.json")
			assert.NoError(t, err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write(resp)
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.Attachments("TEST")
	assert.NoError(t, err)
	assert.Len(t, actual, 2)

	expected := &Attachment{
		ID:       "10700",
		Filename: "screenshot one.png",
		Size:     82976,
		MimeType: "image/png",
		Content:  "https://example.com/secure/attachment/10700/screenshot%20one.png",
		Created:  "2026-09-25T17:11:45.000+0300",
	}
	assert.Equal(t, expected.Filename, actual[0].Filename)
	assert.Equal(t, expected.ID, actual[0].ID)
	assert.Equal(t, expected.Size, actual[0].Size)
	assert.Equal(t, "John Doe", actual[0].Author.DisplayName)

	apiVersion2 = true
	unexpectedStatusCode = true

	_, err = client.AttachmentsV2("TEST")
	assert.Error(t, &ErrUnexpectedResponse{}, err)
}

func TestDownloadAttachment(t *testing.T) {
	var unexpectedStatusCode bool

	content := []byte("attachment-content")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/secure/attachment/10700/screenshot one.png", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		if unexpectedStatusCode {
			w.WriteHeader(404)
			return
		}

		w.WriteHeader(200)
		_, _ = w.Write(content)
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	var buf bytes.Buffer

	err := client.DownloadAttachment(&Attachment{
		ID:       "10700",
		Filename: "screenshot one.png",
	}, &buf)
	assert.NoError(t, err)
	assert.Equal(t, content, buf.Bytes())

	unexpectedStatusCode = true

	err = client.DownloadAttachment(&Attachment{ID: "10700", Filename: "screenshot one.png"}, &buf)
	assert.Error(t, &ErrUnexpectedResponse{}, err)

	err = client.DownloadAttachment(nil, &buf)
	assert.Equal(t, ErrNoResult, err)
}
