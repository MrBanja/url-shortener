package api

import (
	"errors"
	"net/url"
)

type EncodeRequest struct {
	URL string `json:"url"`
}

func (r *EncodeRequest) Validate() error {
	if r.URL == "" {
		return errors.New("url is required")
	}
	if u, err := url.Parse(r.URL); err != nil {
		return errors.New("url is invalid")
	} else if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("url scheme must be http or https")
	} else if u.Host == "" {
		return errors.New("url host is required")
	}
	return nil
}
