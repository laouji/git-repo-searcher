package github_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/laouji/git-repo-searcher/pkg/github"
	"github.com/stretchr/testify/suite"
)

type clientTestSuite struct {
	suite.Suite
	server *httptest.Server

	repos []github.Repository
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
	client := github.NewClient(500*time.Millisecond, server.URL)
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
	client := github.NewClient(500*time.Millisecond, server.URL)
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
	client := github.NewClient(500*time.Millisecond, server.URL)
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
	client := github.NewClient(500*time.Millisecond, server.URL)
	repos, err := client.ListPublicEvents(context.Background(), 50, 1)
	s.NoError(err)
	s.Require().Len(repos, 3)
}
