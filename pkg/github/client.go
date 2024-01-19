package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const headerAccept = "application/vnd.github+json"
const headerAPIVer = "2022-11-28"

type Client struct {
	innerClient *http.Client
	baseURL     string
}

func NewClient(httpClient *http.Client, baseURL string) *Client {
	return &Client{httpClient, baseURL}
}
func (c *Client) do(req *http.Request) (res *http.Response, err error) {
	req.Header.Add("Accept", headerAccept)
	req.Header.Add("X-GitHub-Api-Version", headerAPIVer)
	res, err = c.innerClient.Do(req)
	if err != nil {
		return res, fmt.Errorf("failed to do request: %w", err)
	}
	if res.StatusCode > 399 {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return res, fmt.Errorf("failed to unmarshal error response for status %d: %w", res.StatusCode, err)
		}
		defer res.Body.Close()
		return res, fmt.Errorf("unexpected status of %d for req: %q", res.StatusCode, b)
	}
	return res, nil
}

func (c *Client) ListPublicRepos(ctx context.Context) (repos []Repository, err error) {
	path := "/repositories"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return repos, fmt.Errorf("failed to create request to %q: %w", path, err)
	}
	res, err := c.do(req)
	if err != nil {
		return repos, fmt.Errorf("failed to do request to %q: %w", path, err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return repos, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	defer res.Body.Close()

	err = json.Unmarshal(b, &repos)
	if err != nil {
		return repos, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return repos, nil
}
