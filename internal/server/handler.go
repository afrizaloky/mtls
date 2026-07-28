package server

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
)

const maxBodySize int64 = 65536

type apiError struct {
	Error string `json:"error"`
}

type healthResponse struct {
	Status string `json:"status"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiError{Error: msg})
}

func writeHealth(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(healthResponse{Status: "ok"})
}

type echoHandler struct {
	codec *codec
}

func newEchoHandler(c *codec) *echoHandler {
	return &echoHandler{codec: c}
}

func (h *echoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/echo" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/octet-stream" {
		writeError(w, http.StatusUnsupportedMediaType, "unsupported media type")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	plaintext, err := readPlaintext(r.Body)
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(w, http.StatusRequestEntityTooLarge, "request too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}
	response, err := buildResponse(plaintext, h.codec)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(response)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/health" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeHealth(w)
}

func setupRoutes(h *echoHandler, c *codec) http.Handler {
	_ = c
	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/health", healthHandler)
	mux.HandleFunc("POST /v1/echo", h.ServeHTTP)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/echo" && r.URL.Path != "/v1/health" {
			writeError(w, http.StatusNotFound, "not found")
			return
		}
		mux.ServeHTTP(w, r)
	})
}
