package github_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/laouji/git-repo-searcher/pkg/github"
	"github.com/stretchr/testify/suite"
)

type clientTestSuite struct {
	suite.Suite
	server *httptest.Server

	appID      string
	privateKey *rsa.PrivateKey
	repos      []github.Repository
}

func TestClient(t *testing.T) {
	suite.Run(t, new(clientTestSuite))
}

func (s *clientTestSuite) SetupTest() {
	languageHandlerFn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"Ruby": 644}`))
	})
	s.server = httptest.NewServer(languageHandlerFn)
	s.repos = []github.Repository{
		{Name: "RepoName", LanguagesURL: s.server.URL},
	}
}

func (s *clientTestSuite) TeardownTest() {
	s.server.Close()
}

func (s *clientTestSuite) TestListRepos_Success() {
	expectedSince := int64(77777)
	repoHandlerFn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Equal(fmt.Sprintf("since=%d", expectedSince), r.URL.RawQuery)
		w.WriteHeader(http.StatusOK)
		b, err := json.Marshal(s.repos)
		if err != nil {
			s.Fail(err.Error())
		}
		w.Write(b)
	})
	server := httptest.NewServer(repoHandlerFn)
	client := github.NewClient(500*time.Millisecond, server.URL, s.appID, s.privateKey)
	repos, err := client.ListPublicRepos(context.Background(), expectedSince)
	s.NoError(err)
	s.Require().Len(repos, len(s.repos))
	s.Equal(s.repos[0].Name, repos[0].Name)
	s.Equal(s.server.URL, repos[0].LanguagesURL)
}

func (s *clientTestSuite) TestListRepos_InternalServerErrorResponse() {
	repoHandlerFn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("some error code"))
	})
	server := httptest.NewServer(repoHandlerFn)
	client := github.NewClient(500*time.Millisecond, server.URL, s.appID, s.privateKey)
	_, err := client.ListPublicRepos(context.Background(), int64(2))
	s.Require().Error(err)
	s.Regexp(http.StatusInternalServerError, err.Error())
}

func (s *clientTestSuite) TestListRepos_ContextCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	repoHandlerFn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cancel()
		w.Write([]byte("client hung up before response"))
	})
	server := httptest.NewServer(repoHandlerFn)
	client := github.NewClient(500*time.Millisecond, server.URL, s.appID, s.privateKey)
	_, err := client.ListPublicRepos(ctx, int64(3333))
	s.Require().Error(err)
	s.Regexp("context cancel", err.Error())
}

func (s *clientTestSuite) TestListPublicEvents() {
	repoHandlerFn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal([]github.Event{
			{Type: "CreateEvent"},
			{Type: "PushEvent"},
			{Type: "DeleteEvent"},
		})
		if err != nil {
			s.Fail(err.Error())
		}
		w.Write(b)
	})
	server := httptest.NewServer(repoHandlerFn)
	client := github.NewClient(500*time.Millisecond, server.URL, s.appID, s.privateKey)
	repos, err := client.ListPublicEvents(context.Background(), 50, 1)
	s.NoError(err)
	s.Require().Len(repos, 3)
}

func (s *clientTestSuite) TestSetAccessToken() {
	someToken := "token"
	appID := 7777
	tokenHandlerFn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := []byte(fmt.Sprintf(`{"token":"%s"}`, someToken))
		w.Write(b)
	})
	server := httptest.NewServer(tokenHandlerFn)
	installationsFn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := []byte(fmt.Sprintf(`[{"app_id":%d, "access_tokens_url":"%s"}]`, appID, server.URL+"/app/installations/"))
		w.Write(b)
	})
	server2 := httptest.NewServer(installationsFn)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	s.Require().NoError(err)

	client := github.NewClient(400*time.Millisecond, server2.URL, strconv.Itoa(appID), privateKey)
	tok, err := client.SetAccessToken(context.Background())
	s.NoError(err)
	s.Equal(someToken, tok.Token)
}
