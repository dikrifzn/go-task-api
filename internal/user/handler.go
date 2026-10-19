package user

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	u, err := h.service.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		case errors.Is(err, ErrEmailExists):
			http.Error(w, `{"error":"email already registered"}`, http.StatusConflict)
		default:
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "registration successful",
		"user":    u,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	auth, err := h.service.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			http.Error(w, `{"error":"email or password is incorrect"}`, http.StatusUnauthorized)
		case errors.Is(err, ErrInvalidInput):
			http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		default:
			http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, auth)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Header status sudah dikirim; catat error di aplikasi nyata.
		return
	}
}

// Helper untuk mengambil Bearer token pada langkah middleware berikutnya.
func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "

	if !strings.HasPrefix(header, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(header, prefix))
}
