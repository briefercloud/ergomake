package oauthutils

import (
	"net/http"

	"github.com/gregjones/httpcache"
	"golang.org/x/oauth2"
)

func CachedHTTPClient(token *oauth2.Token, cache httpcache.Cache) *http.Client {
	ts := oauth2.StaticTokenSource(token)
	return &http.Client{
		Transport: &oauth2.Transport{
			Base:   httpcache.NewTransport(cache),
			Source: ts,
		},
	}
}
