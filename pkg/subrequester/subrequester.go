package subrequester

import (
	"context"
	"strings"
	"sync"

	"github.com/laouji/git-repo-searcher/pkg/github"
	"github.com/laouji/git-repo-searcher/pkg/model"
	"github.com/sirupsen/logrus"
)

const (
	FilterKeyLanguage = "language"
)

var PermittedFilterKeys = map[string]struct{}{
	FilterKeyLanguage: {},
}

type SubRequester struct {
	workerCount int

	logger logrus.FieldLogger
	client github.Client
	input  chan github.Repository
	output chan model.Repository
	errs   chan error
}

func NewSubRequester(
	workerCount int,
	client github.Client,
	logger logrus.FieldLogger,
) *SubRequester {
	input := make(chan github.Repository, workerCount)
	output := make(chan model.Repository, workerCount)
	errs := make(chan error, workerCount)
	return &SubRequester{
		workerCount: workerCount,
		client:      client,
		input:       input,
		output:      output,
		errs:        errs,
		logger:      logger,
	}
}

func (s *SubRequester) Collect(
	ctx context.Context,
	in []github.Repository,
	filters map[string]string,
) (out []model.Repository, err error) {
	wg := &sync.WaitGroup{}
	out = make([]model.Repository, 0, len(in))

	// fan out to multiple workers to make time consuming requests
	s.logger.Debugf("spawning %d workers", s.workerCount)
	for i := 0; i <= s.workerCount; i++ {
		wg.Add(1)
		go s.runWorker(ctx, wg, filters)
	}

	done := make(chan struct{})
	go func() {
		for repo := range s.output {
			out = append(out, repo)
		}
		close(done)
	}()

	errsDone := make(chan struct{})
	errs := make([]error, 0, s.workerCount)
	go func() {
		for err := range s.errs {
			s.logger.WithError(err).Error("collect error")
			errs = append(errs, err)
		}
		close(errsDone)
	}()

	for _, repo := range in {
		s.input <- repo
	}
	close(s.input)

	wg.Wait()
	close(s.output)
	close(s.errs)

	<-done
	<-errsDone
	s.logger.Debug("subrequests done")

	if len(errs) > 0 {
		return out, errs[0]
	}
	return out, nil
}

func (s *SubRequester) runWorker(
	ctx context.Context,
	wg *sync.WaitGroup,
	filters map[string]string,
) {
	defer wg.Done()
	for repo := range s.input {
		select {
		case <-ctx.Done():
			s.errs <- ctx.Err()
			break
		default:
			s.fetchSingle(ctx, repo, filters)
		}
	}
}

func (s *SubRequester) fetchSingle(
	ctx context.Context,
	repo github.Repository,
	filters map[string]string,
) {
	attrs, err := s.client.FetchAttribute(ctx, repo.LanguagesURL)
	if err != nil {
		s.errs <- err
	}

	languages := make(map[string]model.Language)
	for k, val := range attrs {
		key := strings.ToLower(k)
		languages[key] = model.Language{Bytes: val}
	}

	relevant := true

	// if a language filter is set check if this entry is relevant
	if langs, ok := filters[FilterKeyLanguage]; ok {
		relevant = false
		for _, l := range strings.Split(strings.ToLower(langs), ",") {
			if _, found := languages[l]; found {
				relevant = true
				break
			}
		}
	}

	// discard any entries that don't match the filters
	if !relevant {
		return
	}

	s.output <- model.Repository{
		FullName:   repo.FullName,
		Owner:      repo.Owner.Login,
		Repository: repo.Name,
		Languages:  languages,
	}
}
