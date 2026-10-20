package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

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
