package server

import (
	"context"
	"time"
)

type pullRequestsService struct {
	repository string
}

type PullRequest struct {
	CreatedAt time.Time
}

func (s *pullRequestsService) List(ctx context.Context) ([]PullRequest, error) {
	return nil, nil
}
