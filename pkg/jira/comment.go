package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ankitpokhrel/jira-cli/pkg/md"
)

// Comment holds issue comment info.
type Comment struct {
	ID           string `json:"id"`
	Author       *User  `json:"author"`
	UpdateAuthor *User  `json:"updateAuthor"`
	Body         string `json:"body"`
	Created      string `json:"created"`
	Updated      string `json:"updated"`
}

const maxCommentsPerPage = 1000

type commentsResponse struct {
	Total    int        `json:"total"`
	Comments []*Comment `json:"comments"`
}

type issueCommentUpdateRequest struct {
	Body string `json:"body"`
}

// GetIssueComments fetches all comments of an issue using v2 version of the GET /issue/{key}/comment endpoint.
func (c *Client) GetIssueComments(key string) ([]*Comment, error) {
	var comments []*Comment

	for startAt := 0; ; {
		path := fmt.Sprintf("/issue/%s/comment?startAt=%d&maxResults=%d", key, startAt, maxCommentsPerPage)

		res, err := c.GetV2(context.Background(), path, Header{
			"Accept":       "application/json",
			"Content-Type": "application/json",
		})
		if err != nil {
			return nil, err
		}
		if res == nil {
			return nil, ErrEmptyResponse
		}
		if res.StatusCode != http.StatusOK {
			_ = res.Body.Close()
			return nil, formatUnexpectedResponse(res)
		}

		var page commentsResponse

		err = json.NewDecoder(res.Body).Decode(&page)
		_ = res.Body.Close()
		if err != nil {
			return nil, err
		}

		comments = append(comments, page.Comments...)

		startAt += len(page.Comments)
		if len(page.Comments) == 0 || startAt >= page.Total {
			return comments, nil
		}
	}
}

// UpdateIssueComment updates a comment of an issue using PUT /issue/{key}/comment/{id} endpoint.
func (c *Client) UpdateIssueComment(key, id, body string) error {
	payload, err := json.Marshal(&issueCommentUpdateRequest{Body: md.ToJiraMD(body)})
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/issue/%s/comment/%s", key, id)
	res, err := c.PutV2(context.Background(), path, payload, Header{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	})
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
	return nil
}

// DeleteIssueComment deletes a comment of an issue using DELETE /issue/{key}/comment/{id} endpoint.
func (c *Client) DeleteIssueComment(key, id string) error {
	path := fmt.Sprintf("/issue/%s/comment/%s", key, id)

	res, err := c.DeleteV2(context.Background(), path, Header{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	})
	if err != nil {
		return err
	}
	if res == nil {
		return ErrEmptyResponse
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusNoContent && res.StatusCode != http.StatusOK {
		return formatUnexpectedResponse(res)
	}
	return nil
}
