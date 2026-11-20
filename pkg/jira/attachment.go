package jira

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const boundaryBytes = 16

var quoteEscaper = strings.NewReplacer(`\`, `\\`, `"`, `\"`)

// Attachment holds issue attachment info.
type Attachment struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
	Content  string `json:"content"`
	Created  string `json:"created"`
	Author   struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	} `json:"author"`
}

type attachmentsResponse struct {
	Fields struct {
		Attachment []*Attachment `json:"attachment"`
	} `json:"fields"`
}

// Attachments fetches attachment metadata of an issue using v3 version of the GET /issue/{key} endpoint.
func (c *Client) Attachments(key string) ([]*Attachment, error) {
	return c.attachments(key, apiVersion3)
}

// AttachmentsV2 fetches attachment metadata of an issue using v2 version of the GET /issue/{key} endpoint.
func (c *Client) AttachmentsV2(key string) ([]*Attachment, error) {
	return c.attachments(key, apiVersion2)
}

func (c *Client) attachments(key, ver string) ([]*Attachment, error) {
	path := fmt.Sprintf("/issue/%s?fields=attachment", key)

	var (
		res *http.Response
		err error
	)

	switch ver {
	case apiVersion2:
		res, err = c.GetV2(context.Background(), path, nil)
	default:
		res, err = c.Get(context.Background(), path, nil)
	}

	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var out attachmentsResponse

	err = json.NewDecoder(res.Body).Decode(&out)

	return out.Fields.Attachment, err
}

// UploadAttachment attaches a local file to an issue using v3 version of the POST /issue/{key}/attachments endpoint.
func (c *Client) UploadAttachment(key, path string) (*Attachment, error) {
	return c.uploadAttachment(key, path, apiVersion3)
}

// UploadAttachmentV2 attaches a local file to an issue using v2 version of the POST /issue/{key}/attachments endpoint.
func (c *Client) UploadAttachmentV2(key, path string) (*Attachment, error) {
	return c.uploadAttachment(key, path, apiVersion2)
}

func (c *Client) uploadAttachment(key, path, ver string) (*Attachment, error) {
	if key == "" {
		return nil, ErrNoResult
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}

	body, size, contentType, err := newAttachmentBody(f, info.Name(), info.Size())
	if err != nil {
		return nil, err
	}

	base := baseURLv3
	if ver == apiVersion2 {
		base = baseURLv2
	}

	endpoint := fmt.Sprintf("%s%s/issue/%s/attachments", c.server, base, url.PathEscape(key))
	headers := Header{
		"Content-Type":      contentType,
		"X-Atlassian-Token": "no-check",
	}

	res, err := c.requestStream(context.Background(), http.MethodPost, endpoint, body, size, headers)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var out []*Attachment

	if err = json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, ErrNoResult
	}

	return out[0], nil
}

// newAttachmentBody builds a multipart body around r with the exact content length,
// so the file can be streamed instead of buffered in memory.
func newAttachmentBody(r io.Reader, filename string, size int64) (io.Reader, int64, string, error) {
	var raw [boundaryBytes]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return nil, 0, "", err
	}

	boundary := hex.EncodeToString(raw[:])
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	head := fmt.Sprintf(
		"--%s\r\nContent-Disposition: form-data; name=\"file\"; filename=\"%s\"\r\nContent-Type: %s\r\n\r\n",
		boundary, escapeQuotes(filename), contentType,
	)
	tail := fmt.Sprintf("\r\n--%s--\r\n", boundary)

	body := io.MultiReader(strings.NewReader(head), r, strings.NewReader(tail))
	length := int64(len(head)) + size + int64(len(tail))

	return body, length, "multipart/form-data; boundary=" + boundary, nil
}

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// DownloadAttachment streams the attachment content into w.
func (c *Client) DownloadAttachment(a *Attachment, w io.Writer) error {
	if a == nil || a.ID == "" {
		return ErrNoResult
	}

	path := fmt.Sprintf("/secure/attachment/%s/%s", a.ID, url.PathEscape(a.Filename))

	res, err := c.request(context.Background(), http.MethodGet, c.server+path, nil, nil)
	if err != nil {
		return err
	}
	if res == nil {
		return ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return formatUnexpectedResponse(res)
	}

	_, err = io.Copy(w, res.Body)

	return err
}
