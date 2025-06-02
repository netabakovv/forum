package usecase_test

import (
	"fmt"

	"github.com/netabakovv/forum/back/forum_service/internal/entities"
	"github.com/netabakovv/forum/back/forum_service/internal/repository/mocks"
	"github.com/netabakovv/forum/back/forum_service/internal/usecase"
	uc_mocks "github.com/netabakovv/forum/back/forum_service/internal/usecase/mocks"
	"github.com/netabakovv/forum/back/pkg/errors"
	pb "github.com/netabakovv/forum/back/proto"

	"context"
	"testing"
	"time"

	"github.com/netabakovv/forum/back/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/mock/gomock"
)

type MockChatRepo struct {
	mock.Mock
}

func (m *MockChatRepo) SaveMessage(ctx context.Context, userID int64, content string) error {
	args := m.Called(ctx, userID, content)
	return args.Error(0)
}

func (m *MockChatRepo) GetMessages(ctx context.Context) ([]*entities.ChatMessage, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.ChatMessage), args.Error(1)
}

func (m *MockChatRepo) DeleteOldMessages(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

// --- Tests ---

func TestChatUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	mockRepo := mocks.NewMockChatRepository(ctrl)
	log := logger.NewStdLogger()
	config := &pb.ChatConfig{
		MaxMessageLength:       100,
		MessageLifetimeMinutes: 60,
	}

	chat := usecase.NewChatUsecase(mockRepo, log, config)

	t.Run("SendMessage - success", func(t *testing.T) {
		msg := &entities.ChatMessage{UserID: 1, Content: "Hello"}

		mockRepo.
			EXPECT().
			SaveMessage(ctx, msg.UserID, msg.Username, msg.Content).
			Return(nil)

		err := chat.SendMessage(ctx, msg)
		assert.NoError(t, err)
	})

	t.Run("SendMessage - too long", func(t *testing.T) {
		long := make([]byte, 101)
		msg := &entities.ChatMessage{UserID: 1, Content: string(long)}

		err := chat.SendMessage(ctx, msg)
		assert.EqualError(t, err, fmt.Sprintf("сообщение слишком длинное (максимум %d символов)", config.MaxMessageLength))
	})

	t.Run("SendMessage - empty", func(t *testing.T) {
		msg := &entities.ChatMessage{UserID: 1, Content: ""}

		err := chat.SendMessage(ctx, msg)
		assert.ErrorIs(t, err, errors.ErrEmptyMessage)
	})

	t.Run("GetMessages - success", func(t *testing.T) {
		expected := []*entities.ChatMessage{
			{ID: 1, UserID: 1, Content: "Hello"},
		}

		mockRepo.
			EXPECT().
			GetMessages(ctx).
			Return(expected, nil)

		result, err := chat.GetMessages(ctx)
		assert.NoError(t, err)
		assert.Equal(t, expected, result)
	})

	t.Run("DeleteOldMessages - success", func(t *testing.T) {
		before := time.Now()

		mockRepo.
			EXPECT().
			DeleteOldMessages(ctx, before).
			Return(nil)

		err := chat.DeleteOldMessages(ctx, before)
		assert.NoError(t, err)
	})
}

func TestSendMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	use := uc_mocks.NewMockChatUsecaseInterface(ctrl)

	t.Run("success", func(t *testing.T) {
		msg := &entities.ChatMessage{UserID: 1, Content: "Hello"}
		use.EXPECT().
			SendMessage(gomock.Any(), msg).
			Return(nil)

		err := use.SendMessage(context.Background(), msg)
		assert.NoError(t, err)
	})

	t.Run("too long message", func(t *testing.T) {
		long := make([]byte, 101)
		msg := &entities.ChatMessage{UserID: 1, Content: string(long)}

		use.EXPECT().
			SendMessage(gomock.Any(), msg).
			Return(errors.ErrMessageTooLong)

		err := use.SendMessage(context.Background(), msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "сообщение слишком длинное")
	})

	t.Run("empty message", func(t *testing.T) {
		msg := &entities.ChatMessage{UserID: 1, Content: ""}

		use.EXPECT().
			SendMessage(gomock.Any(), msg).
			Return(errors.ErrEmptyMessage)

		err := use.SendMessage(context.Background(), msg)
		assert.Equal(t, errors.ErrEmptyMessage, err)
	})
}

func TestGetMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	use := uc_mocks.NewMockChatUsecaseInterface(ctrl)

	expected := []*entities.ChatMessage{
		{ID: 1, UserID: 1, Content: "Hi"},
		{ID: 2, UserID: 2, Content: "Hello"},
	}

	use.EXPECT().
		GetMessages(gomock.Any()).
		Return(expected, nil)

	messages, err := use.GetMessages(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expected, messages)
}

func TestDeleteOldMessages(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	use := uc_mocks.NewMockChatUsecaseInterface(ctrl)

	before := time.Now().Add(-time.Hour)
	use.EXPECT().
		DeleteOldMessages(gomock.Any(), gomock.Any()).
		Return(nil)

	err := use.DeleteOldMessages(context.Background(), before)
	assert.NoError(t, err)
}

func TestDeleteOldMessages_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	use := uc_mocks.NewMockChatUsecaseInterface(ctrl)

	before := time.Now().Add(-time.Hour)
	use.EXPECT().
		DeleteOldMessages(gomock.Any(), gomock.Any()).
		Return(errors.ErrDeleteFailed)

	err := use.DeleteOldMessages(context.Background(), before)
	assert.Error(t, err)
}

func TestCleanupService_Cleanup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chatUC := uc_mocks.NewMockChatUsecaseInterface(ctrl)
	logger := logger.NewStdLogger()

	service := usecase.NewCleanupService(chatUC, logger)

	chatUC.EXPECT().DeleteOldMessages(gomock.Any(), gomock.Any()).Return(nil)

	err := service.Cleanup(10 * time.Minute)
	assert.NoError(t, err)
}

func TestCleanupService_ErrorLogged(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatUC := uc_mocks.NewMockChatUsecaseInterface(ctrl)
	mockLogger := logger.NewStdLogger()

	errExpected := errors.ErrCleanupOldMessage

	mockChatUC.EXPECT().
		DeleteOldMessages(gomock.Any(), gomock.Any()).
		Return(errExpected).
		MinTimes(1)

	service := usecase.NewCleanupService(mockChatUC, mockLogger)

	service.Start(50*time.Millisecond, 1*time.Second)

	time.Sleep(120 * time.Millisecond)
	service.Stop()
	time.Sleep(50 * time.Millisecond)
}

func TestPostUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockPostRepository(ctrl)
	logger := logger.NewStdLogger()
	uc := usecase.NewPostUsecase(repo, logger)

	ctx := context.Background()
	post := &entities.Post{ID: 1, Title: "title", AuthorID: 1}

	t.Run("CreatePost", func(t *testing.T) {
		repo.EXPECT().CreatePost(ctx, post).Return(nil)
		err := uc.CreatePost(ctx, post)
		assert.NoError(t, err)
	})

	t.Run("GetPostByID", func(t *testing.T) {
		repo.EXPECT().GetPostByID(ctx, int64(1)).Return(post, nil)
		res, err := uc.GetPostByID(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, post, res)
	})

	t.Run("UpdatePost", func(t *testing.T) {
		repo.EXPECT().UpdatePost(ctx, post).Return(nil)
		err := uc.UpdatePost(ctx, post)
		assert.NoError(t, err)
	})

	t.Run("DeletePost", func(t *testing.T) {
		repo.EXPECT().DeletePost(ctx, int64(1)).Return(nil)
		err := uc.DeletePost(ctx, 1)
		assert.NoError(t, err)
	})

	t.Run("Posts", func(t *testing.T) {
		repo.EXPECT().Posts(ctx).Return([]*entities.Post{post}, nil)
		res, err := uc.Posts(ctx)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})
}

func TestCommentUsecase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockCommentRepository(ctrl)
	logger := logger.NewStdLogger()
	uc := usecase.NewCommentUsecase(repo, logger)

	ctx := context.Background()
	comment := &entities.Comment{ID: 1, AuthorID: 1, PostID: 2, Content: "text"}

	t.Run("CreateComment", func(t *testing.T) {
		repo.EXPECT().CreateComment(ctx, comment).Return(nil)
		err := uc.CreateComment(ctx, comment)
		assert.NoError(t, err)
	})

	t.Run("GetCommentByID", func(t *testing.T) {
		repo.EXPECT().GetCommentByID(ctx, int64(1)).Return(comment, nil)
		res, err := uc.GetCommentByID(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, comment, res)
	})

	t.Run("GetByPostID", func(t *testing.T) {
		repo.EXPECT().GetByPostID(ctx, int64(2)).Return([]*entities.Comment{comment}, nil)
		res, err := uc.GetByPostID(ctx, 2)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("GetByUserID", func(t *testing.T) {
		repo.EXPECT().GetByUserID(ctx, int64(1)).Return([]*entities.Comment{comment}, nil)
		res, err := uc.GetByUserID(ctx, 1)
		assert.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("UpdateComment", func(t *testing.T) {
		repo.EXPECT().UpdateComment(ctx, comment).Return(nil)
		err := uc.UpdateComment(ctx, comment)
		assert.NoError(t, err)
	})

	t.Run("DeleteComment", func(t *testing.T) {
		repo.EXPECT().DeleteComment(ctx, int64(1)).Return(nil)
		err := uc.DeleteComment(ctx, 1)
		assert.NoError(t, err)
	})
}
