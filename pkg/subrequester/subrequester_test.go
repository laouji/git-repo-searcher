package subrequester_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Scalingo/go-utils/logger"
	"github.com/laouji/git-repo-searcher/pkg/github"
	mock_github "github.com/laouji/git-repo-searcher/pkg/github/mocks"
	"github.com/laouji/git-repo-searcher/pkg/subrequester"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type subRequesterTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	clientMock *mock_github.MockClient
}

func TestSubRequester(t *testing.T) {
	suite.Run(t, new(subRequesterTestSuite))
}

func (s *subRequesterTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.clientMock = mock_github.NewMockClient(s.ctrl)
}

func (s *subRequesterTestSuite) TestCollect_NoFilters() {
	repos := []github.Repository{
		{FullName: "RepoFName1", Name: "RepoName1", LanguagesURL: "http://url.com/1"},
		{FullName: "RepoFName2", Name: "RepoName2", LanguagesURL: "http://url.com/2"},
		{FullName: "RepoFName3", Name: "RepoName3", LanguagesURL: "http://url.com/3"},
		{FullName: "RepoFName4", Name: "RepoName4", LanguagesURL: "http://url.com/4"},
	}
	subRequester := subrequester.NewSubRequester(3, s.clientMock, logger.Default())

	sampleAttribute := map[string]int64{}
	s.clientMock.EXPECT().FetchAttribute(gomock.Any(), repos[0].LanguagesURL).Return(sampleAttribute, nil)
	s.clientMock.EXPECT().FetchAttribute(gomock.Any(), repos[1].LanguagesURL).Return(sampleAttribute, nil)
	s.clientMock.EXPECT().FetchAttribute(gomock.Any(), repos[2].LanguagesURL).Return(sampleAttribute, nil)
	s.clientMock.EXPECT().FetchAttribute(gomock.Any(), repos[3].LanguagesURL).Return(sampleAttribute, nil)

	out, err := subRequester.Collect(context.Background(), repos, map[string]string{})
	s.NoError(err)
	s.Require().Len(out, len(repos))
}

func (s *subRequesterTestSuite) TestCollect_APIErrors() {
	expectedURL := "expectedURL"
	subRequester := subrequester.NewSubRequester(3, s.clientMock, logger.Default())
	repos := []github.Repository{
		{FullName: "RepoFName1", Name: "RepoName1", LanguagesURL: expectedURL},
	}

	expectedErr := errors.New("error")
	s.clientMock.EXPECT().FetchAttribute(gomock.Any(), expectedURL).Return(map[string]int64{}, expectedErr)
	_, err := subRequester.Collect(context.Background(), repos, map[string]string{})
	s.Require().Error(err)
	s.Equal(expectedErr, err)
}

func (s *subRequesterTestSuite) TestCollect_FilterByLanguage() {
	repos := []github.Repository{
		{FullName: "RepoFName1", Name: "RepoName1", LanguagesURL: "http://url.com/1"},
		{FullName: "RepoFName2", Name: "RepoName2", LanguagesURL: "http://url.com/2"},
		{FullName: "RepoFName3", Name: "RepoName3", LanguagesURL: "http://url.com/3"},
		{FullName: "RepoFName4", Name: "RepoName4", LanguagesURL: "http://url.com/4"},
	}
	subRequester := subrequester.NewSubRequester(3, s.clientMock, logger.Default())

	s.clientMock.EXPECT().FetchAttribute(gomock.Any(), repos[0].LanguagesURL).Return(map[string]int64{
		"Ruby":       3434,
		"Javascript": 229,
	}, nil)
	s.clientMock.EXPECT().FetchAttribute(gomock.Any(), repos[1].LanguagesURL).Return(map[string]int64{
		"Go": 799,
	}, nil)
	s.clientMock.EXPECT().FetchAttribute(gomock.Any(), repos[2].LanguagesURL).Return(map[string]int64{
		"Javascript": 333,
		"CSS":        1223,
	}, nil)
	s.clientMock.EXPECT().FetchAttribute(gomock.Any(), repos[3].LanguagesURL).Return(map[string]int64{
		"Rust": 333,
	}, nil)

	filters := map[string]string{"language": "JavaScript"}
	out, err := subRequester.Collect(context.Background(), repos, filters)
	s.NoError(err)
	s.Require().Len(out, 2)
}
