package usecase

import "back/internal/repository"

type CommentUsecase struct {
	repo repository.CommentRepository
}
