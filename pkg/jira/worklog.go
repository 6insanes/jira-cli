package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/ankitpokhrel/jira-cli/pkg/md"
)

// Worklog holds issue worklog info.
type Worklog struct {
	ID               string `json:"id"`
	IssueID          string `json:"issueId"`
	Author           *User  `json:"author"`
	UpdateAuthor     *User  `json:"updateAuthor"`
	Comment          string `json:"comment"`
	Started          string `json:"started"`
	TimeSpent        string `json:"timeSpent"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
}

type worklogsResponse struct {
	Worklogs []*Worklog `json:"worklogs"`
}

type issueWorklogUpdateRequest struct {
	Started   string  `json:"started,omitempty"`
	TimeSpent string  `json:"timeSpent,omitempty"`
	Comment   *string `json:"comment,omitempty"`
}

// GetIssueWorklogs fetches worklogs of an issue using v2 version of the GET /issue/{key}/worklog endpoint.
func (c *Client) GetIssueWorklogs(key string) ([]*Worklog, error) {
	path := fmt.Sprintf("/issue/%s/worklog", key)

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
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, formatUnexpectedResponse(res)
	}

	var out worklogsResponse

	err = json.NewDecoder(res.Body).Decode(&out)

	return out.Worklogs, err
}

// UpdateIssueWorklog updates a worklog of an issue using PUT /issue/{key}/worklog/{id} endpoint.
// Leave a param empty to keep the current value. Pass a comment pointer to update or clear the comment.
func (c *Client) UpdateIssueWorklog(key, id, started, timeSpent string, comment *string, newEstimate string) error {
	worklogReq := issueWorklogUpdateRequest{
		Started:   started,
		TimeSpent: timeSpent,
	}
	if comment != nil {
		converted := md.ToJiraMD(*comment)
		worklogReq.Comment = &converted
	}

	body, err := json.Marshal(&worklogReq)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/issue/%s/worklog/%s", key, id)
	if newEstimate != "" {
		path = fmt.Sprintf("%s?%s", path, worklogEstimateQuery(newEstimate))
	}

	res, err := c.PutV2(context.Background(), path, body, Header{
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

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return formatUnexpectedResponse(res)
	}
	return nil
}

// DeleteIssueWorklog deletes a worklog of an issue using DELETE /issue/{key}/worklog/{id} endpoint.
// Leave the newEstimate param empty to let Jira adjust the remaining estimate automatically.
func (c *Client) DeleteIssueWorklog(key, id, newEstimate string) error {
	path := fmt.Sprintf("/issue/%s/worklog/%s", key, id)
	if newEstimate != "" {
		path = fmt.Sprintf("%s?%s", path, worklogEstimateQuery(newEstimate))
	}

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

func worklogEstimateQuery(newEstimate string) string {
	return url.Values{
		"adjustEstimate": []string{"new"},
		"newEstimate":    []string{newEstimate},
	}.Encode()
}
