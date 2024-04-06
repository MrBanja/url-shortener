package app

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/mrbanja/url-shortener/api"
	"github.com/mrbanja/url-shortener/dal"
)

func Run(ctx context.Context, opt *Options) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))

	repo := dal.New()
	svc := api.New(repo, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /encode", svc.Encode)
	mux.HandleFunc("GET /decode", svc.Decode)
	mux.HandleFunc("GET /health", func(_ http.ResponseWriter, _ *http.Request) {})

	server := &http.Server{
		Addr:    opt.Addr,
		Handler: mux,

		BaseContext: func(listener net.Listener) context.Context { return ctx },
	}

	if err := listenAndServe(ctx, server, logger, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGHUP); err != nil {
		logger.Error("server run errors", "error", err)
	}

	return nil
}

func listenAndServe(
	ctx context.Context,
	server *http.Server,
	logger *slog.Logger,
	signals ...os.Signal,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var (
		wg       sync.WaitGroup
		errorsCh = make(chan error, 2)
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		select {
		case <-ctx.Done():
			logger.Info("start server: ctx canceled")
			return
		default:
		}

		logger.Info("server listen and serve", slog.String("Addr", server.Addr))
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server listen and serve", "error", err)
			errorsCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		c := make(chan os.Signal)
		signal.Notify(c, signals...)

		select {
		case <-ctx.Done():
		case sig := <-c:
			logger.Warn("received signal", slog.String("signal", sig.String()))
		}

		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		logger.Warn("shutting down server")
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("server shutdown", "error", err)
			errorsCh <- err
		}
		logger.Warn("shut down")
	}()

	wg.Wait()
	close(errorsCh)
	var err error
	for e := range errorsCh {
		err = errors.Join(err, e)
	}

	return err
}
