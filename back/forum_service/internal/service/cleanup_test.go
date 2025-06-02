package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/netabakovv/forum/back/forum_service/internal/service"
	mock_usecase "github.com/netabakovv/forum/back/forum_service/internal/usecase/mocks"
	"github.com/netabakovv/forum/back/pkg/logger"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCleanupService_CleanupOldMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUC := mock_usecase.NewMockChatUsecaseInterface(ctrl)
	log := logger.NewStdLogger()

	cleanup := service.NewCleanupService(mockChatUC, log)

	t.Run("success", func(t *testing.T) {
		lifetime := 30 * time.Minute
		cutoff := time.Now().Add(-lifetime)

		mockChatUC.
			EXPECT().
			DeleteOldMessages(gomock.Any(), gomock.AssignableToTypeOf(cutoff)).
			Return(nil)

		err := cleanup.CleanupOldMessages(lifetime)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		lifetime := 30 * time.Minute
		cutoff := time.Now().Add(-lifetime)

		mockChatUC.
			EXPECT().
			DeleteOldMessages(gomock.Any(), gomock.AssignableToTypeOf(cutoff)).
			Return(errors.New("db error"))

		err := cleanup.CleanupOldMessages(lifetime)
		assert.EqualError(t, err, "db error")
	})
}

func TestCleanupService_StartAndStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUC := mock_usecase.NewMockChatUsecaseInterface(ctrl)
	log := logger.NewStdLogger()

	cleanup := service.NewCleanupService(mockChatUC, log)

	lifetime := 1 * time.Millisecond
	interval := 2 * time.Millisecond

	mockChatUC.
		EXPECT().
		DeleteOldMessages(gomock.Any(), gomock.Any()).
		MinTimes(1)

	cleanup.Start(interval, lifetime)
	time.Sleep(5 * time.Millisecond)
	cleanup.Stop()
}
