package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recover(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, q *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			logger.Error("http handler panicked", "panic", rec)
			debug.PrintStack()

			err, ok := rec.(error)
			if !ok {
				return
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, q)
	})
}
