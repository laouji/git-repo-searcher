package github

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	headerAccept = "application/vnd.github+json"
	headerAPIVer = "2022-11-28"
)

var (
	ErrAuthentication  = errors.New("github API returned 401")
	ErrRateLimit       = errors.New("github API rate limit reached")
	ErrAppNotInstalled = errors.New("github app is not installed")

	reInstallation = regexp.MustCompile("^/app/installations")
)

type Client interface {
	SetAccessToken(ctx context.Context) (Token, error)
	ListPublicRepos(ctx context.Context, since int64) (repos []Repository, err error)
	ListPublicEvents(ctx context.Context, limit, offset int) (events []Event, err error)
	FetchAttribute(ctx context.Context, url string) (attributes map[string]int64, err error)
}

type client struct {
	innerClient     *http.Client
	baseURL         string
	appID           string
	privateKey      *rsa.PrivateKey
	canAuthenticate bool

	mu              sync.RWMutex
	accessToken     *Token
	tokenRefreshURL string
}

func NewClient(
	clientTimeout time.Duration,
	baseURL string,
	githubAppID string,
	githubPrivateKey *rsa.PrivateKey,
) Client {
	var canAuthenticate bool
	if githubAppID != "" && githubPrivateKey != nil {
		canAuthenticate = true
	}
	return &client{
		innerClient:     &http.Client{Timeout: clientTimeout},
		baseURL:         baseURL,
		canAuthenticate: canAuthenticate,
		appID:           githubAppID,
		privateKey:      githubPrivateKey,

		mu: sync.RWMutex{},
	}
}

func (c *client) generateJWT() (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": now.Add(5 * time.Minute).Unix(),
		"iss": c.appID,
		"alg": "RS256",
		//"alg": jwt.SigningMethodRS256.Alg(),
	})
	return token.SignedString(c.privateKey)
}

func (c *client) addAuthHeader(req *http.Request) (err error) {
	if c.canAuthenticate == false {
		return nil
	}

	// installation paths need Json Web tokens
	if reInstallation.MatchString(req.URL.Path) {
		jwToken, err := c.generateJWT()
		if err != nil {
			return fmt.Errorf("failed to generate json web token for app %q: %w", c.appID, err)
		}
		req.Header.Add("Authorization", "Bearer "+jwToken)
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.accessToken == nil {
		return fmt.Errorf("token was not set in client")
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken.Token)
	return nil
}

func (c *client) do(req *http.Request) (res *http.Response, err error) {
	if err := c.addAuthHeader(req); err != nil {
		return res, fmt.Errorf("failed to add auth header for app %q: %w", c.appID, err)
	}

	req.Header.Add("Accept", headerAccept)
	req.Header.Add("X-GitHub-Api-Version", headerAPIVer)
	res, err = c.innerClient.Do(req)
	if err != nil {
		return res, fmt.Errorf("failed to do request: %w", err)
	}
	if res.StatusCode > 399 {
		switch res.StatusCode {
		//		case http.StatusUnauthorized:
		//			return res, ErrAuthentication
		case http.StatusForbidden:
			return res, ErrRateLimit
		}

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

func (c *client) SetAccessToken(ctx context.Context) (token Token, err error) {
	if !c.canAuthenticate {
		return token, nil
	}

	if c.tokenRefreshURL == "" {
		err := c.setTokenRefreshURL(ctx)
		if err != nil {
			return token, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.tokenRefreshURL, nil)
	if err != nil {
		return token, fmt.Errorf("failed to create request to %s: %w", c.tokenRefreshURL, err)
	}
	res, err := c.do(req)
	if err != nil {
		return token, fmt.Errorf("failed to do request to %s: %w", c.tokenRefreshURL, err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return token, fmt.Errorf("failed to unmarshal response from %s: %w", c.tokenRefreshURL, err)
	}
	defer res.Body.Close()

	err = json.Unmarshal(b, &token)
	if err != nil {
		return token, fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = &token
	return token, nil
}

func (c *client) setTokenRefreshURL(ctx context.Context) error {
	path := "/app/installations"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request to %s: %w", path, err)
	}
	res, err := c.do(req)
	if err != nil {
		return fmt.Errorf("failed to do request to %s: %w", path, err)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response from %s: %w", path, err)
	}
	defer res.Body.Close()

	var installations []Installation
	err = json.Unmarshal(b, &installations)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	for _, installation := range installations {
		if strconv.Itoa(installation.AppID) == c.appID {
			c.tokenRefreshURL = installation.AccessTokensURL
			return nil
		}
	}
	return ErrAppNotInstalled
}
