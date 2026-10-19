package task

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidInput = errors.New("invalid task input")

type Store interface {
	Create(context.Context, int64, CreateRequest) (Task, error)
	GetAll(context.Context, int64) ([]Task, error)
	GetByID(context.Context, int64, int64) (Task, error)
	Update(context.Context, int64, int64, UpdateRequest) (Task, error)
	Delete(context.Context, int64, int64) error

	List(context.Context, int64, ListOptions) (ListResponse, error)
}

type Service struct {
	repo Store
}

func NewService(repo Store) *Service {
	return &Service{repo: repo}
}

func validateTask(title, description string) (string, string, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)

	if title == "" || len(title) > 200 {
		return "", "", ErrInvalidInput
	}

	if len(description) > 2000 {
		return "", "", ErrInvalidInput
	}

	return title, description, nil
}

func (s *Service) Create(
	ctx context.Context,
	userID int64,
	req CreateRequest,
) (Task, error) {
	title, description, err := validateTask(req.Title, req.Description)
	if err != nil {
		return Task{}, err
	}

	req.Title = title
	req.Description = description

	return s.repo.Create(ctx, userID, req)
}

func (s *Service) List(
	ctx context.Context,
	userID int64,
	options ListOptions,
) (ListResponse, error) {
	if options.Page < 1 {
		options.Page = 1
	}

	if options.Limit < 1 {
		options.Limit = 10
	}

	if options.Limit > 100 {
		options.Limit = 100
	}

	options.Search = strings.TrimSpace(options.Search)
	if len(options.Search) > 100 {
		options.Search = options.Search[:100]
	}

	if options.Sort != "oldest" {
		options.Sort = "newest"
	}

	return s.repo.List(ctx, userID, options)
}

func (s *Service) GetAll(
	ctx context.Context,
	userID int64,
) ([]Task, error) {
	return s.repo.GetAll(ctx, userID)
}

func (s *Service) GetByID(
	ctx context.Context,
	userID, taskID int64,
) (Task, error) {
	return s.repo.GetByID(ctx, userID, taskID)
}

func (s *Service) Update(
	ctx context.Context,
	userID, taskID int64,
	req UpdateRequest,
) (Task, error) {
	title, description, err := validateTask(req.Title, req.Description)
	if err != nil {
		return Task{}, err
	}

	req.Title = title
	req.Description = description

	return s.repo.Update(ctx, userID, taskID, req)
}

func (s *Service) Delete(
	ctx context.Context,
	userID, taskID int64,
) error {
	return s.repo.Delete(ctx, userID, taskID)
}
