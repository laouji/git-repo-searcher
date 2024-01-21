package authentication_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Scalingo/go-utils/logger"
	"github.com/laouji/git-repo-searcher/pkg/authentication"
	"github.com/laouji/git-repo-searcher/pkg/github"
	mock_github "github.com/laouji/git-repo-searcher/pkg/github/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type authenticatorTestSuite struct {
	suite.Suite
	ctrl       *gomock.Controller
	clientMock *mock_github.MockClient
}

func TestAuthenticator(t *testing.T) {
	suite.Run(t, new(authenticatorTestSuite))
}

func (s *authenticatorTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.clientMock = mock_github.NewMockClient(s.ctrl)
}

func (s *authenticatorTestSuite) TestAuthenticate_BailOnAuthError() {
	interval := time.Second
	threshold := 500 * time.Millisecond
	authenticator := authentication.NewAuthenticator(s.clientMock, logger.Default())

	expectedErr := errors.New("some err")
	token := github.Token{ExpiresAt: time.Now().Add(time.Hour)}
	s.clientMock.EXPECT().SetAccessToken(gomock.Any()).Return(token, expectedErr)

	err := authenticator.Authenticate(context.Background(), interval, threshold)
	s.Require().Error(err)
	s.Equal(expectedErr, err)
}
