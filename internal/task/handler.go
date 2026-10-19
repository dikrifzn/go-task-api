package task

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go-task-api/internal/user"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	t, err := h.service.Create(r.Context(), u.ID, req)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid task input")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := r.URL.Query()

	opts := ListOptions{
		Page:   1,
		Limit:  10,
		Search: query.Get("search"),
		Sort:   query.Get("sort"),
	}

	if query.Has("page") {
		page, err := strconv.Atoi(query.Get("page"))
		if err != nil || page < 1 {
			writeError(w, http.StatusBadRequest, "invalid page")
			return
		}
		opts.Page = page
	}

	if query.Has("limit") {
		limit, err := strconv.Atoi(query.Get("limit"))
		if err != nil || limit < 1 {
			writeError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		opts.Limit = limit
	}

	if query.Has("completed") {
		completed, err := strconv.ParseBool(query.Get("completed"))
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid completed value")
			return
		}
		opts.Completed = &completed
	}

	result, err := h.service.List(r.Context(), u.ID, opts)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, "invalid query parameters")
		} else {
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := parseTaskID(w, r)
	if !ok {
		return
	}

	t, err := h.service.GetByID(r.Context(), u.ID, id)
	if err != nil {
		h.handleTaskError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := parseTaskID(w, r)
	if !ok {
		return
	}

	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	t, err := h.service.Update(r.Context(), u.ID, id, req)
	if err != nil {
		h.handleTaskError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	u, ok := user.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, ok := parseTaskID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), u.ID, id); err != nil {
		h.handleTaskError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleTaskError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	if errors.Is(err, ErrInvalidInput) {
		writeError(w, http.StatusBadRequest, "invalid task input")
		return
	}

	writeError(w, http.StatusInternalServerError, "internal server error")
}

func parseTaskID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid task ID")
		return 0, false
	}

	return id, true
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
