package grpc_test

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/netabakovv/forum/back/forum_service/internal/delivery/grpc"
	"github.com/netabakovv/forum/back/forum_service/internal/entities"
	mock_usecase "github.com/netabakovv/forum/back/forum_service/internal/usecase/mocks"
	pb "github.com/netabakovv/forum/back/proto"
	"github.com/stretchr/testify/assert"
)

func TestCreatePost_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postUC := mock_usecase.NewMockPostUsecaseInterface(ctrl)
	commentUC := mock_usecase.NewMockCommentUsecaseInterface(ctrl)
	chatUC := mock_usecase.NewMockChatUsecaseInterface(ctrl)

	server := grpc.NewForumServer(nil, postUC, commentUC, chatUC)

	req := &pb.CreatePostRequest{
		Title:          "Test Title",
		Content:        "Test Content",
		AuthorId:       1,
		AuthorUsername: "user",
	}

	postUC.EXPECT().CreatePost(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, post *entities.Post) error {
			post.ID = 42
			return nil
		})

	resp, err := server.CreatePost(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(42), resp.Post.Id)
	assert.Equal(t, req.Title, resp.Post.Title)
	assert.Equal(t, req.Content, resp.Post.Content)
}

func TestCreatePost_InvalidInput(t *testing.T) {
	server := grpc.NewForumServer(nil, nil, nil, nil)

	req := &pb.CreatePostRequest{
		Title:   "",
		Content: "",
	}

	resp, err := server.CreatePost(context.Background(), req)
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestGetPost_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postUC := mock_usecase.NewMockPostUsecaseInterface(ctrl)
	server := grpc.NewForumServer(nil, postUC, nil, nil)

	postUC.EXPECT().GetPostByID(gomock.Any(), int64(10)).Return(&entities.Post{
		ID:           10,
		Title:        "Title",
		Content:      "Content",
		AuthorID:     1,
		AuthorName:   "Author",
		CreatedAt:    time.Now(),
		CommentCount: 2,
	}, nil)

	req := &pb.GetPostRequest{PostId: 10}
	resp, err := server.GetPost(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int64(10), resp.Post.Id)
	assert.Equal(t, "Title", resp.Post.Title)
}

func TestForumServer_GetByPostID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mock_usecase.NewMockPostUsecaseInterface(ctrl)
	mockCommentUC := mock_usecase.NewMockCommentUsecaseInterface(ctrl)
	mockChatUC := mock_usecase.NewMockChatUsecaseInterface(ctrl)

	srv := grpc.NewForumServer(nil, mockPostUC, mockCommentUC, mockChatUC)

	ctx := context.Background()
	created := time.Now()

	mockComment := &entities.Comment{
		ID:         1,
		PostID:     2,
		AuthorID:   3,
		AuthorName: "john",
		Content:    "hi",
		CreatedAt:  created,
	}

	mockCommentUC.EXPECT().
		GetByPostID(ctx, int64(2)).
		Return([]*entities.Comment{mockComment}, nil)

	req := &pb.GetCommentsByPostIDRequest{PostId: 2}
	resp, err := srv.GetByPostID(ctx, req)
	require.NoError(t, err)
	require.Len(t, resp.Comments, 1)
	require.Equal(t, "hi", resp.Comments[0].Content)
}

func TestUpdatePost_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postUC := mock_usecase.NewMockPostUsecaseInterface(ctrl)
	server := grpc.NewForumServer(nil, postUC, nil, nil)

	title := "Updated Title"
	content := "Updated Content"

	postUC.EXPECT().UpdatePost(gomock.Any(), gomock.Any()).Return(nil)
	postUC.EXPECT().GetPostByID(gomock.Any(), int64(1)).Return(&entities.Post{
		ID:           1,
		Title:        title,
		Content:      content,
		AuthorID:     1,
		AuthorName:   "user",
		CreatedAt:    time.Now(),
		CommentCount: 0,
	}, nil)

	req := &pb.UpdatePostRequest{
		PostId:  1,
		Title:   &title,
		Content: &content,
	}

	resp, err := server.UpdatePost(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, title, resp.Post.Title)
}

func TestDeletePost_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postUC := mock_usecase.NewMockPostUsecaseInterface(ctrl)
	server := grpc.NewForumServer(nil, postUC, nil, nil)

	postUC.EXPECT().DeletePost(gomock.Any(), int64(42)).Return(errors.New("db error"))

	resp, err := server.DeletePost(context.Background(), &pb.DeletePostRequest{PostId: 42})
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestForumServer_Posts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postUC := mock_usecase.NewMockPostUsecaseInterface(ctrl)
	srv := grpc.NewForumServer(nil, postUC, nil, nil)

	ctx := context.Background()
	created := time.Now()
	mockPost := &entities.Post{
		ID:           1,
		Title:        "Test",
		Content:      "Content",
		AuthorID:     42,
		AuthorName:   "tester",
		CreatedAt:    created,
		CommentCount: 3,
	}

	postUC.EXPECT().Posts(ctx).Return([]*entities.Post{mockPost}, nil)

	resp, err := srv.Posts(ctx, &pb.ListPostsRequest{})
	require.NoError(t, err)
	require.Len(t, resp.Posts, 1)
	require.Equal(t, int64(1), resp.Posts[0].Id)
	require.Equal(t, "Test", resp.Posts[0].Title)
}

func TestCreateComment_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	commentUC := mock_usecase.NewMockCommentUsecaseInterface(ctrl)
	server := grpc.NewForumServer(nil, nil, commentUC, nil)

	commentUC.EXPECT().CreateComment(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, comment *entities.Comment) error {
			comment.ID = 77
			comment.CreatedAt = time.Now()
			return nil
		})

	req := &pb.CreateCommentRequest{
		PostId:         1,
		Content:        "Nice post!",
		AuthorId:       2,
		AuthorUsername: "tester",
	}

	resp, err := server.CreateComment(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, int64(77), resp.Comment.Id)
	assert.Equal(t, "Nice post!", resp.Comment.Content)
}

func TestGetCommentByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	commentUC := mock_usecase.NewMockCommentUsecaseInterface(ctrl)
	server := grpc.NewForumServer(nil, nil, commentUC, nil)

	commentUC.EXPECT().GetCommentByID(gomock.Any(), int64(404)).Return(nil, errors.New("not found"))

	resp, err := server.GetCommentByID(context.Background(), &pb.GetCommentRequest{CommentId: 404})
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestForumServer_GetCommentByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mock_usecase.NewMockPostUsecaseInterface(ctrl)
	mockCommentUC := mock_usecase.NewMockCommentUsecaseInterface(ctrl)
	mockChatUC := mock_usecase.NewMockChatUsecaseInterface(ctrl)

	srv := grpc.NewForumServer(nil, mockPostUC, mockCommentUC, mockChatUC)

	ctx := context.Background()
	created := time.Now()

	mockComment := &entities.Comment{
		ID:         1,
		PostID:     5,
		AuthorID:   7,
		AuthorName: "admin",
		Content:    "comment content",
		CreatedAt:  created,
	}

	mockCommentUC.EXPECT().
		GetCommentByID(ctx, int64(1)).
		Return(mockComment, nil)

	req := &pb.GetCommentRequest{CommentId: 1}
	resp, err := srv.GetCommentByID(ctx, req)
	require.NoError(t, err)
	require.Equal(t, mockComment.Content, resp.Comment.Content)
	require.Equal(t, mockComment.AuthorName, resp.Comment.AuthorUsername)
}

func TestForumServer_Comments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mock_usecase.NewMockPostUsecaseInterface(ctrl)
	mockCommentUC := mock_usecase.NewMockCommentUsecaseInterface(ctrl)
	mockChatUC := mock_usecase.NewMockChatUsecaseInterface(ctrl)

	srv := grpc.NewForumServer(nil, mockPostUC, mockCommentUC, mockChatUC)

	ctx := context.Background()
	created := time.Now()

	mockComment := &entities.Comment{
		ID:         10,
		PostID:     99,
		AuthorID:   777,
		AuthorName: "someuser",
		Content:    "text",
		CreatedAt:  created,
	}

	mockCommentUC.EXPECT().
		GetByPostID(ctx, int64(99)).
		Return([]*entities.Comment{mockComment}, nil)

	req := &pb.ListCommentsRequest{PostId: 99}
	resp, err := srv.Comments(ctx, req)
	require.NoError(t, err)
	require.Len(t, resp.Comments, 1)
	require.Equal(t, "text", resp.Comments[0].Content)
	require.Equal(t, int64(10), resp.Comments[0].Id)
}

func TestForumServer_UpdateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mock_usecase.NewMockPostUsecaseInterface(ctrl)
	mockCommentUC := mock_usecase.NewMockCommentUsecaseInterface(ctrl)
	mockChatUC := mock_usecase.NewMockChatUsecaseInterface(ctrl)

	srv := grpc.NewForumServer(nil, mockPostUC, mockCommentUC, mockChatUC)

	ctx := context.Background()
	now := time.Now()

	comment := &entities.Comment{
		ID:         1,
		PostID:     2,
		AuthorID:   3,
		AuthorName: "bob",
		Content:    "updated",
		CreatedAt:  now,
	}

	mockCommentUC.EXPECT().
		UpdateComment(ctx, gomock.AssignableToTypeOf(&entities.Comment{})).
		DoAndReturn(func(_ context.Context, c *entities.Comment) error {
			// симулируем, что usecase подставляет остальные поля
			c.AuthorID = comment.AuthorID
			c.AuthorName = comment.AuthorName
			c.PostID = comment.PostID
			c.CreatedAt = comment.CreatedAt
			return nil
		})

	content := "updated"
	req := &pb.UpdateCommentRequest{
		CommentId: 1,
		Content:   &content,
	}
	resp, err := srv.UpdateComment(ctx, req)
	require.NoError(t, err)
	require.Equal(t, "updated", resp.Comment.Content)
	require.Equal(t, "bob", resp.Comment.AuthorUsername)
}

func TestSendMessage_EmptyContent(t *testing.T) {
	server := grpc.NewForumServer(nil, nil, nil, nil)

	resp, err := server.SendMessage(context.Background(), &pb.ChatMessage{
		Content: "",
	})
	assert.Nil(t, resp)
	assert.Error(t, err)
}

func TestForumServer_SendMessage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostUC := mock_usecase.NewMockPostUsecaseInterface(ctrl)
	mockCommentUC := mock_usecase.NewMockCommentUsecaseInterface(ctrl)
	mockChatUC := mock_usecase.NewMockChatUsecaseInterface(ctrl)

	srv := grpc.NewForumServer(nil, mockPostUC, mockCommentUC, mockChatUC)

	t.Run("успешная отправка сообщения", func(t *testing.T) {
		ctx := context.Background()
		req := &pb.ChatMessage{
			UserId:  42,
			Content: "привет, мир!",
		}

		mockChatUC.EXPECT().
			SendMessage(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, msg *entities.ChatMessage) error {
				require.Equal(t, req.UserId, msg.UserID)
				require.Equal(t, req.Content, msg.Content)
				require.WithinDuration(t, time.Now(), msg.CreatedAt, time.Second)
				return nil
			})

		resp, err := srv.SendMessage(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("пустое сообщение — ошибка", func(t *testing.T) {
		ctx := context.Background()
		req := &pb.ChatMessage{
			UserId:  42,
			Content: "",
		}

		resp, err := srv.SendMessage(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		s, _ := status.FromError(err)
		require.Equal(t, codes.InvalidArgument, s.Code())
		require.Contains(t, s.Message(), "содержание сообщения обязательно")
	})

	t.Run("ошибка при отправке в usecase", func(t *testing.T) {
		ctx := context.Background()
		req := &pb.ChatMessage{
			UserId:  42,
			Content: "fail me",
		}

		mockChatUC.EXPECT().
			SendMessage(ctx, gomock.Any()).
			Return(errors.New("db is down"))

		resp, err := srv.SendMessage(ctx, req)
		require.Error(t, err)
		require.Nil(t, resp)
		s, _ := status.FromError(err)
		require.Equal(t, codes.Internal, s.Code())
		require.Contains(t, s.Message(), "не удалось отправить сообщение")
	})
}

func TestGetMessages_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chatUC := mock_usecase.NewMockChatUsecaseInterface(ctrl)
	server := grpc.NewForumServer(nil, nil, nil, chatUC)

	chatUC.EXPECT().GetMessages(gomock.Any()).Return([]*entities.ChatMessage{
		{UserID: 1, Content: "Hi", CreatedAt: time.Now()},
	}, nil)

	resp, err := server.GetMessages(context.Background(), &pb.GetMessagesRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Messages, 1)
}
