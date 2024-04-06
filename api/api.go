package api

import (
	"log/slog"
	"net/http"

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

}

func (a *API) Decode(w http.ResponseWriter, r *http.Request) {

}
