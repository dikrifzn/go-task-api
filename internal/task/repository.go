package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

var ErrNotFound = errors.New("task not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(
	ctx context.Context,
	userID int64,
	req CreateRequest,
) (Task, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO tasks (user_id, title, description)
		VALUES (?, ?, ?)
	`, userID, req.Title, req.Description)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Task{}, fmt.Errorf("get task ID: %w", err)
	}

	return r.GetByID(ctx, userID, id)
}

func (r *Repository) GetAll(
	ctx context.Context,
	userID int64,
) ([]Task, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, title, description, completed, created_at
		FROM tasks
		WHERE user_id = ?
		ORDER BY id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)

	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Title,
			&t.Description,
			&t.Completed,
			&t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (r *Repository) GetByID(
	ctx context.Context,
	userID, taskID int64,
) (Task, error) {
	var t Task

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, title, description, completed, created_at
		FROM tasks
		WHERE id = ? AND user_id = ?
	`, taskID, userID).Scan(
		&t.ID,
		&t.UserID,
		&t.Title,
		&t.Description,
		&t.Completed,
		&t.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, fmt.Errorf("get task: %w", err)
	}

	return t, nil
}

func (r *Repository) Update(
	ctx context.Context,
	userID, taskID int64,
	req UpdateRequest,
) (Task, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tasks
		SET title = ?, description = ?, completed = ?
		WHERE id = ? AND user_id = ?
	`, req.Title, req.Description, req.Completed, taskID, userID)
	if err != nil {
		return Task{}, fmt.Errorf("update task: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return Task{}, fmt.Errorf("check updated task: %w", err)
	}
	if n == 0 {
		return Task{}, ErrNotFound
	}

	return r.GetByID(ctx, userID, taskID)
}

func (r *Repository) Delete(
	ctx context.Context,
	userID, taskID int64,
) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM tasks
		WHERE id = ? AND user_id = ?
	`, taskID, userID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted task: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *Repository) List(
	ctx context.Context,
	userID int64,
	opts ListOptions,
) (ListResponse, error) {
	conditions := []string{"user_id = ?"}
	args := []any{userID}

	if opts.Completed != nil {
		conditions = append(conditions, "completed = ?")
		args = append(args, *opts.Completed)
	}

	if opts.Search != "" {
		conditions = append(conditions, "LOWER(title) LIKE ?")
		args = append(args, "%"+strings.ToLower(opts.Search)+"%")
	}

	where := strings.Join(conditions, " AND ")

	var total int64
	countQuery := "SELECT COUNT(*) FROM tasks WHERE " + where

	if err := r.db.QueryRowContext(
		ctx, countQuery, args...,
	).Scan(&total); err != nil {
		return ListResponse{}, fmt.Errorf("count tasks: %w", err)
	}

	orderBy := "id DESC"
	if opts.Sort == "oldest" {
		orderBy = "id ASC"
	}

	query := `
		SELECT id, user_id, title, description, completed, created_at
		FROM tasks
		WHERE ` + where + `
		ORDER BY ` + orderBy + `
		LIMIT ? OFFSET ?
	`

	offset := (opts.Page - 1) * opts.Limit
	queryArgs := append(append([]any{}, args...), opts.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return ListResponse{}, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]Task, 0)

	for rows.Next() {
		var t Task

		if err := rows.Scan(
			&t.ID,
			&t.UserID,
			&t.Title,
			&t.Description,
			&t.Completed,
			&t.CreatedAt,
		); err != nil {
			return ListResponse{}, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return ListResponse{}, fmt.Errorf("iterate tasks: %w", err)
	}

	var totalPages int64
	if total > 0 {
		totalPages = (total + int64(opts.Limit) - 1) / int64(opts.Limit)
	}

	return ListResponse{
		Data:       tasks,
		Page:       opts.Page,
		Limit:      opts.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}
