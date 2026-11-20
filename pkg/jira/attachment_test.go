package jira

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestUploadAttachment(t *testing.T) {
	const filePerm = 0o600

	var (
		apiVersion2          bool
		unexpectedStatusCode bool
		emptyResponse        bool

		gotName    string
		gotContent []byte
	)

	dir := t.TempDir()
	path := filepath.Join(dir, "screenshot one.png")

	content := []byte("attachment-content")

	assert.NoError(t, os.WriteFile(path, content, filePerm))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiVersion2 {
			assert.Equal(t, "/rest/api/2/issue/TEST/attachments", r.URL.Path)
		} else {
			assert.Equal(t, "/rest/api/3/issue/TEST/attachments", r.URL.Path)
		}

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "no-check", r.Header.Get("X-Atlassian-Token"))

		f, hdr, err := r.FormFile("file")
		assert.NoError(t, err)

		if err == nil {
			defer func() { _ = f.Close() }()

			gotName = hdr.Filename
			gotContent, err = io.ReadAll(f)
			assert.NoError(t, err)
		}

		switch {
		case unexpectedStatusCode:
			w.WriteHeader(400)
		case emptyResponse:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte("[]"))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`[{"id":"10701","filename":"screenshot one.png","size":18,"mimeType":"image/png"}]`))
		}
	}))
	defer server.Close()

	client := NewClient(Config{Server: server.URL}, WithTimeout(3*time.Second))

	actual, err := client.UploadAttachment("TEST", path)
	assert.NoError(t, err)
	assert.Equal(t, "10701", actual.ID)
	assert.Equal(t, "screenshot one.png", actual.Filename)
	assert.Equal(t, int64(18), actual.Size)
	assert.Equal(t, "screenshot one.png", gotName)
	assert.Equal(t, content, gotContent)

	emptyResponse = true

	_, err = client.UploadAttachment("TEST", path)
	assert.Equal(t, ErrNoResult, err)

	emptyResponse = false
	apiVersion2 = true
	unexpectedStatusCode = true

	_, err = client.UploadAttachmentV2("TEST", path)
	assert.Error(t, &ErrUnexpectedResponse{}, err)

	_, err = client.UploadAttachment("", path)
	assert.Equal(t, ErrNoResult, err)

	_, err = client.UploadAttachment("TEST", filepath.Join(dir, "missing.png"))
	assert.Error(t, err)

	_, err = client.UploadAttachment("TEST", dir)
	assert.Error(t, err)
}
