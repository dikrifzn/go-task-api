package task

import "time"

type Task struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"-"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type UpdateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

type ListOptions struct {
	Page      int
	Limit     int
	Completed *bool
	Search    string
	Sort      string
}

type ListResponse struct {
	Data       []Task `json:"data"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	Total      int64  `json:"total"`
	TotalPages int64  `json:"total_pages"`
}
