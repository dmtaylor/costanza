package util

import "net/http"

func Healthcheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(200)
}
