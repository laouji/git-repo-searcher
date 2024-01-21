package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Scalingo/go-handlers"
	"github.com/Scalingo/go-utils/logger"
	"github.com/laouji/git-repo-searcher/pkg/github"
	"github.com/laouji/git-repo-searcher/pkg/handler"
	"github.com/laouji/git-repo-searcher/pkg/searcher"
	"github.com/laouji/git-repo-searcher/pkg/subrequester"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logger.Default()
	log.Info("Initializing app")
	cfg, err := newConfig()
	if err != nil {
		log.WithError(err).Error("Failed to initialize configuration")
		os.Exit(1)
	}

	if err := RunServer(cfg, log); err != nil {
		log.WithError(err).Error("Failed to run web server")
		os.Exit(2)
	}
}

func RunServer(cfg *Config, log logrus.FieldLogger) error {
	githubClient := github.NewClient(cfg.ClientTimeout, cfg.GithubURL)
	repoSearcher := searcher.NewSearcher(githubClient)
	subRequester := subrequester.NewSubRequester(cfg.WorkerCount, githubClient, log)

	log.Info("Initializing routes")
	router := handlers.NewRouter(log)
	router.HandleFunc("/ping", handler.Pong)
	// Initialize web server and configure the following routes:
	router.HandleFunc("/repos", handler.Repos(repoSearcher, subRequester))
	// GET /stats

	log = log.WithField("port", cfg.Port)
	log.Info("Listening...")
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), router)
	if err != nil {
		return fmt.Errorf("Failed to listen to the given port: %d", cfg.Port)
	}
	return nil
}
