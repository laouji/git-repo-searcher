package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/Scalingo/go-handlers"
	"github.com/Scalingo/go-utils/logger"
	"github.com/golang-jwt/jwt"
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

	var privateKey *rsa.PrivateKey
	if cfg.GithubAppID != "" {
		privateKey, err = readAPIPrivateKey(cfg.GithubPrivateKey)
		if err != nil {
			log.WithError(err).Error("Failed to read file: %s", cfg.GithubPrivateKey)
			os.Exit(2)
		}
	}

	// TODO handle signals
	ctx := context.Background()
	if err := RunServer(ctx, cfg, log, privateKey); err != nil {
		log.WithError(err).Error("Failed to run web server")
		os.Exit(2)
	}
}

func readAPIPrivateKey(path string) (key *rsa.PrivateKey, err error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open private key file: %w", err)
	}
	pemBytes, err := ioutil.ReadAll(fh)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}
	return jwt.ParseRSAPrivateKeyFromPEM(pemBytes)
}

func RunServer(
	ctx context.Context,
	cfg *Config,
	log logrus.FieldLogger,
	privateKey *rsa.PrivateKey,
) error {
	githubClient := github.NewClient(cfg.ClientTimeout, cfg.GithubURL, cfg.GithubAppID, privateKey)

	if err := Authenticate(ctx, githubClient, log); err != nil {
		return fmt.Errorf("failed to authenticate with github API: %w", err)
	}

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

func Authenticate(ctx context.Context, githubClient github.Client, log logrus.FieldLogger) error {
	token, err := githubClient.SetAccessToken(ctx)
	if err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return

			case <-ticker.C:
				threshold := time.Now().UTC().Add(10 * time.Minute)
				if threshold.After(token.ExpiresAt) {
					log.Debug("refreshing github access token")
					token, err = githubClient.SetAccessToken(ctx)
					if err != nil {
						log.WithError(err).Error("failed to reset access token")
					}
				}
			}
		}
	}()
	return nil
}
