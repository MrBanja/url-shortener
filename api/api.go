package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/rs/xid"

	"github.com/mrbanja/url-shortener/dal"
)

type API struct {
	repo   *dal.DAL
	logger *slog.Logger
}

func New(repo *dal.DAL, logger *slog.Logger) *API {
	return &API{
		repo:   repo,
		logger: logger.With("module", "api"),
	}
}

func (a *API) Encode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req EncodeRequest
	defer func() { _ = r.Body.Close() }()
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	short := xid.New()
	if err := a.repo.Set(ctx, short.String(), req.URL); err != nil {
		a.logger.Error("set long url", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write([]byte(short.String())); err != nil {
		a.logger.Error("encode response", "error", err)
	}
}

func (a *API) Decode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	short := r.URL.Query().Get("short")
	if short == "" {
		http.Error(w, "short query param is required", http.StatusBadRequest)
		return
	}

	long, err := a.repo.Get(ctx, short)
	if err != nil {
		if errors.Is(err, dal.NotFoundError) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		a.logger.Error("get long url", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, long, http.StatusFound)
}
