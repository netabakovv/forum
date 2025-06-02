package service

import (
	"context"
	"time"

	"github.com/netabakovv/forum/back/forum_service/internal/usecase"
	"github.com/netabakovv/forum/back/pkg/logger"
)

type CleanupServiceInterface interface {
	NewCleanupService(chatUC *usecase.ChatUsecase, logger logger.Logger) *CleanupService
	Start(interval, lifetime time.Duration)
	Stop()
	cleanupOldMessages(lifetime time.Duration) error
}

type CleanupService struct {
	chatUC usecase.ChatUsecaseInterface
	logger logger.Logger
	stop   chan struct{}
}

func NewCleanupService(chatUC usecase.ChatUsecaseInterface, logger logger.Logger) *CleanupService {
	return &CleanupService{
		chatUC: chatUC,
		logger: logger,
		stop:   make(chan struct{}),
	}
}

func (s *CleanupService) Start(interval, lifetime time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := s.CleanupOldMessages(lifetime); err != nil {
					s.logger.Error("cleanup failed",
						logger.NewField("error", err))
				}
			case <-s.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *CleanupService) Stop() {
	s.stop <- struct{}{}
}

func (s *CleanupService) CleanupOldMessages(lifetime time.Duration) error {
	cutoff := time.Now().Add(-lifetime)
	return s.chatUC.DeleteOldMessages(context.Background(), cutoff)
}
