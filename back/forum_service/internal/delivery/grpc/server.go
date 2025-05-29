package grpc

import (
	"back/forum_service/internal/entities"
	"back/forum_service/internal/usecase"
	pb "back/proto"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ForumServer struct {
	pb.UnimplementedForumServiceServer
	authService pb.AuthServiceClient
	postUC      *usecase.PostUsecase
	commentUC   *usecase.CommentUsecase
	chatUC      *usecase.ChatUsecase
}

// NewForumServer — конструктор (удобно для внедрения зависимостей)
func NewForumServer(
	authService pb.AuthServiceClient,
	postUC *usecase.PostUsecase,
	commentUC *usecase.CommentUsecase,
	chatUC *usecase.ChatUsecase,
) *ForumServer {
	return &ForumServer{
		authService: authService,
		postUC:      postUC,
		commentUC:   commentUC,
		chatUC:      chatUC,
	}
}

// Post operations
func (s *ForumServer) CreatePost(ctx context.Context, req *pb.CreatePostRequest) (*pb.PostResponse, error) {
	if req.Title == "" || req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "заголовок и содержание обязательны")
	}

	post := &entities.Post{
		Title:        req.Title,
		Content:      req.Content,
		AuthorID:     req.AuthorId,
		AuthorName:   req.AuthorUsername,
		CreatedAt:    time.Now(),
		CommentCount: 0,
	}

	err := s.postUC.CreatePost(ctx, post)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось создать пост")
	}

	return &pb.PostResponse{
		Post: &pb.Post{
			Id:             post.ID,
			Title:          post.Title,
			Content:        post.Content,
			AuthorId:       post.AuthorID,
			AuthorUsername: post.AuthorName,
			CreatedAt:      post.CreatedAt.Unix(),
			CommentCount:   post.CommentCount,
		},
	}, nil
}

func (s *ForumServer) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.PostResponse, error) {
	if req.PostId == 0 {
		return nil, status.Error(codes.InvalidArgument, "идентификатор поста обязателен")
	}

	post, err := s.postUC.GetPostByID(ctx, req.PostId)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось получить пост")
	}

	return &pb.PostResponse{
		Post: &pb.Post{
			Id:             post.ID,
			Title:          post.Title,
			Content:        post.Content,
			AuthorId:       post.AuthorID,
			AuthorUsername: post.AuthorName,
			CreatedAt:      post.CreatedAt.Unix(),
			CommentCount:   post.CommentCount,
		},
	}, nil
}

func (s *ForumServer) GetByPostID(ctx context.Context, req *pb.GetCommentsByPostIDRequest) (*pb.ListCommentsResponse, error) {
	if req.PostId == 0 {
		return nil, status.Error(codes.InvalidArgument, "идентификатор поста обязателен")
	}
	comments, err := s.commentUC.GetByPostID(ctx, req.PostId)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось получить комментарии")
	}
	protoComments := make([]*pb.Comment, 0, len(comments))
	for _, c := range comments {
		protoComments = append(protoComments, &pb.Comment{
			Id:             c.ID,
			PostId:         c.PostID,
			AuthorId:       c.AuthorID,
			AuthorUsername: c.AuthorName,
			Content:        c.Content,
			CreatedAt:      c.CreatedAt.Unix(),
		})
	}
	return &pb.ListCommentsResponse{
		Comments:   protoComments,
		TotalCount: int32(len(comments)),
	}, nil
}

func (s *ForumServer) UpdatePost(ctx context.Context, req *pb.UpdatePostRequest) (*pb.PostResponse, error) {
	if req.PostId == 0 {
		return nil, status.Error(codes.InvalidArgument, "идентификатор поста обязателен")
	}

	post := &entities.Post{
		ID: req.PostId,
	}

	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Content != nil {
		post.Content = *req.Content
	}

	err := s.postUC.UpdatePost(ctx, post)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось обновить пост")
	}

	updatedPost, err := s.postUC.GetPostByID(ctx, req.PostId)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось получить обновлённый пост")
	}

	return &pb.PostResponse{
		Post: &pb.Post{
			Id:             updatedPost.ID,
			Title:          updatedPost.Title,
			Content:        updatedPost.Content,
			AuthorId:       updatedPost.AuthorID,
			AuthorUsername: updatedPost.AuthorName,
			CreatedAt:      updatedPost.CreatedAt.Unix(),
			CommentCount:   updatedPost.CommentCount,
		},
	}, nil
}

func (s *ForumServer) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.EmptyMessage, error) {
	if req.PostId == 0 {
		return nil, status.Error(codes.InvalidArgument, "идентификатор поста обязателен")
	}
	err := s.postUC.DeletePost(ctx, req.PostId)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось удалить пост")
	}
	return &pb.EmptyMessage{}, nil
}

func (s *ForumServer) Posts(ctx context.Context, req *pb.ListPostsRequest) (*pb.ListPostsResponse, error) {
	posts, err := s.postUC.Posts(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось получить список постов")
	}

	pbPosts := make([]*pb.Post, len(posts))
	for i, post := range posts {
		pbPosts[i] = &pb.Post{
			Id:             post.ID,
			Title:          post.Title,
			Content:        post.Content,
			AuthorId:       post.AuthorID,
			AuthorUsername: post.AuthorName,
			CreatedAt:      post.CreatedAt.Unix(),
			CommentCount:   post.CommentCount,
		}
	}

	return &pb.ListPostsResponse{
		Posts: pbPosts,
	}, nil
}

// Comment operations
func (s *ForumServer) CreateComment(ctx context.Context, req *pb.CreateCommentRequest) (*pb.CommentResponse, error) {
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "содержание комментария обязательно")
	}

	comment := &entities.Comment{
		Content:    req.Content,
		AuthorID:   req.AuthorId,
		PostID:     req.PostId,
		AuthorName: req.AuthorUsername,
	}

	err := s.commentUC.CreateComment(ctx, comment)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("не удалось создать комментарий %w", err))
	}

	return &pb.CommentResponse{
		Comment: &pb.Comment{
			Id:             comment.ID,
			Content:        comment.Content,
			AuthorId:       comment.AuthorID,
			AuthorUsername: comment.AuthorName,
			PostId:         comment.PostID,
			CreatedAt:      comment.CreatedAt.Unix(),
		},
	}, nil
}

func (s *ForumServer) GetCommentByID(ctx context.Context, req *pb.GetCommentRequest) (*pb.CommentResponse, error) {
	if req.CommentId == 0 {
		return nil, status.Error(codes.InvalidArgument, "идентификатор комментария обязателен")
	}
	comment, err := s.commentUC.GetCommentByID(ctx, req.CommentId)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось получить комментарий")
	}
	return &pb.CommentResponse{
		Comment: &pb.Comment{
			Id:             comment.ID,
			Content:        comment.Content,
			AuthorId:       comment.AuthorID,
			AuthorUsername: comment.AuthorName,
			PostId:         comment.PostID,
			CreatedAt:      comment.CreatedAt.Unix(),
		},
	}, nil
}

func (s *ForumServer) Comments(ctx context.Context, req *pb.ListCommentsRequest) (*pb.ListCommentsResponse, error) {
	if req.PostId == 0 {
		return nil, status.Error(codes.InvalidArgument, "идентификатор поста обязателен")
	}

	comments, err := s.commentUC.GetByPostID(ctx, req.PostId)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось получить комментарии")
	}

	pbComments := make([]*pb.Comment, len(comments))
	for i, comment := range comments {
		pbComments[i] = &pb.Comment{
			Id:             comment.ID,
			Content:        comment.Content,
			AuthorId:       comment.AuthorID,
			AuthorUsername: comment.AuthorName,
			PostId:         comment.PostID,
			CreatedAt:      comment.CreatedAt.Unix(),
		}
	}

	return &pb.ListCommentsResponse{
		Comments: pbComments,
	}, nil
}

func (s *ForumServer) UpdateComment(ctx context.Context, req *pb.UpdateCommentRequest) (*pb.CommentResponse, error) {
	if req.CommentId == 0 {
		return nil, status.Error(codes.InvalidArgument, "идентификатор комментария обязателен")
	}

	comment := &entities.Comment{
		ID:      req.CommentId,
		Content: req.GetContent(),
	}

	err := s.commentUC.UpdateComment(ctx, comment)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось обновить комментарий")
	}

	return &pb.CommentResponse{
		Comment: &pb.Comment{
			Id:             comment.ID,
			Content:        comment.Content,
			AuthorId:       comment.AuthorID,
			AuthorUsername: comment.AuthorName,
			PostId:         comment.PostID,
			CreatedAt:      comment.CreatedAt.Unix(),
		},
	}, nil
}

func (s *ForumServer) DeleteComment(ctx context.Context, req *pb.DeleteCommentRequest) (*pb.EmptyMessage, error) {
	fmt.Sprintln("YA TUT")
	if req.CommentId == 0 {
		return nil, status.Error(codes.InvalidArgument, "идентификатор комментария и пользователя обязательны")
	}

	err := s.commentUC.DeleteComment(ctx, req.CommentId)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось удалить комментарий")
	}

	return &pb.EmptyMessage{}, nil
}

// Chat operations
func (s *ForumServer) SendMessage(ctx context.Context, req *pb.ChatMessage) (*pb.EmptyMessage, error) {
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "содержание сообщения обязательно")
	}

	msg := &entities.ChatMessage{
		UserID:    req.UserId,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	err := s.chatUC.SendMessage(ctx, msg)
	if err != nil {
		return nil, status.Error(codes.Internal, "не удалось отправить сообщение")
	}

	return &pb.EmptyMessage{}, nil
}

func (s *ForumServer) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	messages, err := s.chatUC.GetMessages(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("не удалось получить сообщения %w", err))
	}

	pbMessages := make([]*pb.ChatMessage, len(messages))
	for i, msg := range messages {
		pbMessages[i] = &pb.ChatMessage{
			UserId:    msg.UserID,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Unix(),
		}
	}

	return &pb.GetMessagesResponse{
		Messages: pbMessages,
	}, nil
}
