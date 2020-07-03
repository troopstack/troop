package utils

import (
	"net/http"
)

func HttpHandler(h *http.Request) *http.Request {
	h.Header.Add("Http-Token", Config().General.Token)
	return h
}
