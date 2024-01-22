package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Scalingo/go-handlers"
	"github.com/Scalingo/go-utils/logger"
	"github.com/golang-jwt/jwt"
	"github.com/laouji/git-repo-searcher/pkg/authentication"
	"github.com/laouji/git-repo-searcher/pkg/github"
	"github.com/laouji/git-repo-searcher/pkg/handler"
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
			log.WithError(err).Errorf("Failed to read file: %s", cfg.GithubPrivateKey)
			os.Exit(2)
		}
	}

	appCtx, cancel := context.WithCancel(context.Background())
	server, err := configureServer(appCtx, cfg, log, privateKey)
	if err != nil {
		log.WithError(err).Error("Failed to configure web server")
		os.Exit(2)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Info("Listening...")
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Errorf("Failed to listen to the given port: %d", cfg.Port)
			os.Exit(2)
		}
		log.Info("Web server stopped receiving new conns")
	}()

	// block until signal is received
	<-signalCh
	cancel() // stop authenticator loop

	shutdownCtx, release := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer release()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Failed to gracefully shutdown server")
		os.Exit(2)
	}
	log.Info("Graceful shutdown complete")
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

func configureServer(
	ctx context.Context,
	cfg *Config,
	log logrus.FieldLogger,
	privateKey *rsa.PrivateKey,
) (*http.Server, error) {
	githubClient := github.NewClient(cfg.ClientTimeout, cfg.GithubURL, cfg.GithubAppID, privateKey)
	authenticator := authentication.NewAuthenticator(githubClient, log)

	if err := authenticator.Authenticate(ctx, cfg.AuthInterval, cfg.AuthRefreshBuffer); err != nil {
		return nil, fmt.Errorf("failed to authenticate with github API: %w", err)
	}

	log.Info("Initializing routes")
	router := handlers.NewRouter(log)
	router.HandleFunc("/ping", handler.Pong)
	// Initialize web server and configure the following routes:
	router.HandleFunc("/repos", handler.Repos(log, githubClient, cfg.WorkerCount))

	log = log.WithField("port", cfg.Port)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}
	return server, nil
}
