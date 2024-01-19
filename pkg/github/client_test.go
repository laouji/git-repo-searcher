package github_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/pkg/github"
	"github.com/stretchr/testify/suite"
)

type clientTestSuite struct {
	suite.Suite
	server     *httptest.Server
	httpClient *http.Client

	repos []github.Repository
}

func TestClient(t *testing.T) {
	suite.Run(t, new(clientTestSuite))
}

func (s *clientTestSuite) SetupTest() {
	s.httpClient = &http.Client{Timeout: 500 * time.Millisecond}
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
	repoHandlerFn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		b, err := json.Marshal(s.repos)
		if err != nil {
			s.Fail(err.Error())
		}
		w.Write(b)
	})
	server := httptest.NewServer(repoHandlerFn)
	client := github.NewClient(s.httpClient, server.URL)
	repos, err := client.ListPublicRepos(context.Background())
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
	client := github.NewClient(s.httpClient, server.URL)
	_, err := client.ListPublicRepos(context.Background())
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
	client := github.NewClient(s.httpClient, server.URL)
	_, err := client.ListPublicRepos(ctx)
	s.Require().Error(err)
	s.Regexp("context cancel", err.Error())
}
