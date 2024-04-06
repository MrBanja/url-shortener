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

	"github.com/rs/cors"

	"github.com/mrbanja/url-shortener/api"
	"github.com/mrbanja/url-shortener/api/middleware"
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

	repo := dal.MustNew(ctx, opt.RedisConnStr)
	defer func() { _ = repo.Close() }()
	svc := api.New(repo, logger)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /encode", svc.Encode)
	mux.HandleFunc("GET /decode", svc.Decode)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := repo.Ping(r.Context()); err != nil {
			logger.Error("health check", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	})

	handler := middleware.Recover(mux, logger)
	handler = cors.Default().Handler(handler)

	server := &http.Server{
		Addr:    opt.Addr,
		Handler: handler,

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
