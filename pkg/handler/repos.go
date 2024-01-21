package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/Scalingo/go-handlers"
	"github.com/Scalingo/go-utils/logger"
	"github.com/laouji/git-repo-searcher/pkg/github"
	"github.com/laouji/git-repo-searcher/pkg/searcher"
	"github.com/laouji/git-repo-searcher/pkg/subrequester"
	"github.com/sirupsen/logrus"
)

func Repos(
	log logrus.FieldLogger,
	githubClient github.Client,
	workerCount int,
) handlers.HandlerFunc {
	repoSearcher := searcher.NewSearcher(githubClient)
	return func(w http.ResponseWriter, r *http.Request, vars map[string]string) error {
		log := logger.Get(r.Context())

		repos, err := repoSearcher.Search(r.Context())
		if err != nil {
			errorResponse(w, log, http.StatusInternalServerError, err)
			return err
		}

		subRequester := subrequester.NewSubRequester(workerCount, githubClient, log)
		out, err := subRequester.Collect(r.Context(), repos, filters(r))
		if err != nil {
			errorResponse(w, log, http.StatusInternalServerError, err)
			return err
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		err = json.NewEncoder(w).Encode(out)
		if err != nil {
			log.WithError(err).Error("Failed to encode repos response JSON")
			return err
		}
		return nil
	}
}

func filters(r *http.Request) map[string]string {
	filters := make(map[string]string)

	queryParams := r.URL.Query()
	for key, _ := range subrequester.PermittedFilterKeys {
		k := strings.ToLower(key)
		if val := queryParams.Get(k); len(val) > 0 {
			filters[k] = val
		}
	}
	return filters
}

func errorResponse(w http.ResponseWriter, log logrus.FieldLogger, status int, err error) {
	var msg string
	switch {
	case errors.Is(err, github.ErrAuthentication):
		status = http.StatusUnauthorized
		msg = github.ErrAuthentication.Error()
	case errors.Is(err, github.ErrRateLimit):
		status = http.StatusForbidden
		msg = github.ErrRateLimit.Error()
	default:
		msg = "internal error"
	}

	log.WithError(err).Error("Request failed")
	w.WriteHeader(status)
	var body = struct {
		Message string `json:"message"`
	}{Message: msg}
	err = json.NewEncoder(w).Encode(body)
	if err != nil {
		w.Write([]byte("failed to encode response body"))
	}
}
