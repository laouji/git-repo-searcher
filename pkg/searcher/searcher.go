package searcher

import (
	"context"
	"fmt"

	"github.com/laouji/git-repo-searcher/pkg/github"
)

const (
	expectedResults = 100
	wantedType      = "CreateEvent"
	wantedRefType   = "repository"
)

type Searcher struct {
	client github.Client
}

func NewSearcher(client github.Client) *Searcher {
	return &Searcher{
		client: client,
	}
}

func (s *Searcher) Search(ctx context.Context) (out []github.Repository, err error) {
	lastID, err := s.lastRepoID(ctx)
	if err != nil {
		return out, fmt.Errorf("failed to fetch last repo ID: %w", err)
	}
	return s.client.ListPublicRepos(ctx, lastID-int64(expectedResults))
}

func (s *Searcher) lastRepoID(ctx context.Context) (ID int64, err error) {
	page := 0
	for {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("aborted searching for repos: %w", ctx.Err())
		default:
			recentEvents, err := s.client.ListPublicEvents(ctx, expectedResults, page)
			if err != nil {
				return ID, fmt.Errorf("failed to list public events: %w", err)
			}

			// if we didn't find a valid Repo ID then we increment the page and keep looking back
			ID = s.validate(ctx, recentEvents)
			if ID > 0 {
				return ID, nil
			}
			page++
		}
	}
	return 0, nil
}

func (s *Searcher) validate(ctx context.Context, events []github.Event) (ID int64) {
	for _, event := range events {
		if event.Type != wantedType {
			continue
		}
		if event.Payload.RefType == wantedRefType {
			return event.Repo.ID
		}
	}
	return 0
}
