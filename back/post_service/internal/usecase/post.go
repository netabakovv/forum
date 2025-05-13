package usecase

import "back/internal/repository"

type PostUsecase struct {
	repo repository.PostRepository
}
