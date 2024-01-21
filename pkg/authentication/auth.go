package authentication

import (
	"context"
	"time"

	"github.com/laouji/git-repo-searcher/pkg/github"
	"github.com/sirupsen/logrus"
)

type Authenticator struct {
	githubClient github.Client
	logger       logrus.FieldLogger
}

func NewAuthenticator(githubClient github.Client, log logrus.FieldLogger) *Authenticator {
	return &Authenticator{githubClient, log}
}

func (a *Authenticator) Authenticate(ctx context.Context, interval, buffer time.Duration) error {
	token, err := a.githubClient.SetAccessToken(ctx)
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return

			case <-ticker.C:
				cutoff := time.Now().UTC().Add(buffer)
				if cutoff.After(token.ExpiresAt) {
					a.logger.Debug("refreshing github access token")
					token, err = a.githubClient.SetAccessToken(ctx)
					if err != nil {
						a.logger.WithError(err).Error("failed to reset access token")
					}
				}
			}
		}
	}()
	return nil
}
