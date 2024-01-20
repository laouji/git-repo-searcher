package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const headerAccept = "application/vnd.github+json"
const headerAPIVer = "2022-11-28"

type Client interface {
	ListPublicRepos(ctx context.Context, since int64) (repos []Repository, err error)
	ListPublicEvents(ctx context.Context, limit, offset int) (events []Event, err error)
	FetchAttribute(ctx context.Context, url string) (attributes map[string]int64, err error)
}

type client struct {
	innerClient *http.Client
	baseURL     string
}

func NewClient(httpClient *http.Client, baseURL string) Client {
	return &client{httpClient, baseURL}
}
func (c *client) do(req *http.Request) (res *http.Response, err error) {
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

// ListPublicRepos makes a request to the Github API to fetch repos with IDs higher than since
// https://docs.github.com/en/rest/repos/repos?apiVersion=2022-11-28#list-public-repositories
func (c *client) ListPublicRepos(
	ctx context.Context,
	since int64,
) (repos []Repository, err error) {
	path := "/repositories"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return repos, fmt.Errorf("failed to create request to %s: %w", path, err)
	}

	q := req.URL.Query()
	q.Add("since", strconv.FormatInt(since, 10))
	req.URL.RawQuery = q.Encode()

	res, err := c.do(req)
	if err != nil {
		return repos, fmt.Errorf("failed to do request to %s: %w", path, err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return repos, fmt.Errorf("failed to read repos response body: %w", err)
	}
	defer res.Body.Close()

	err = json.Unmarshal(b, &repos)
	if err != nil {
		return repos, fmt.Errorf("failed to unmarshal repos response: %w", err)
	}
	return repos, nil
}

// ListPublicEvents finds recent events exposed by github
// https://docs.github.com/en/rest/activity/events?apiVersion=2022-11-28#list-public-events
func (c *client) ListPublicEvents(
	ctx context.Context,
	limit int,
	offset int,
) (events []Event, err error) {
	path := "/events"
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return events, fmt.Errorf("failed to parse url %s%s: %w", c.baseURL, path, err)
	}

	q := u.Query()
	q.Add("per_page", fmt.Sprintf("%d", limit))
	q.Add("page", fmt.Sprintf("%d", offset))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return events, fmt.Errorf("failed to create request to %s: %w", path, err)
	}

	res, err := c.do(req)
	if err != nil {
		return events, fmt.Errorf("failed to do request to %s: %w", path, err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return events, fmt.Errorf("failed to read events response body: %w", err)
	}
	defer res.Body.Close()

	err = json.Unmarshal(b, &events)
	if err != nil {
		return events, fmt.Errorf("failed to unmarshal events response: %w", err)
	}
	return events, nil
}

// FetchAttribute can be used to fetch sub-attributes linked in the Repositories API response
// it expects a json body with unpredictable json keys and plots them into a map
func (c *client) FetchAttribute(ctx context.Context, url string) (attributes map[string]int64, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return attributes, fmt.Errorf("failed to create request to %s: %w", url, err)
	}
	res, err := c.do(req)
	if err != nil {
		return attributes, fmt.Errorf("failed to do request to %s: %w", url, err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return attributes, fmt.Errorf("failed to unmarshal response from %s: %w", url, err)
	}
	defer res.Body.Close()

	err = json.Unmarshal(b, &attributes)
	if err != nil {
		return attributes, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return attributes, nil
}
