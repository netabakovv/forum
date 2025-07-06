package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/netabakovv/forum/back/forum_service/internal/entities"
	"github.com/netabakovv/forum/back/forum_service/internal/repository"
	"github.com/netabakovv/forum/back/pkg/errors"
	"github.com/netabakovv/forum/back/pkg/logger"
	pb "github.com/netabakovv/forum/back/proto"
)

type ChatUsecaseInterface interface {
	DeleteOldMessages(ctx context.Context, cutoff time.Time) error
	GetMessages(ctx context.Context) ([]*entities.ChatMessage, error)
	SendMessage(ctx context.Context, msg *entities.ChatMessage) error
}

type ChatUsecase struct {
	repo            repository.ChatRepository
	logger          logger.Logger
	maxMessageLen   int
	messageLifetime time.Duration
}

func NewChatUsecase(repo repository.ChatRepository, logger logger.Logger, config *pb.ChatConfig) *ChatUsecase {
	return &ChatUsecase{
		repo:            repo,
		logger:          logger,
		maxMessageLen:   int(config.MaxMessageLength),
		messageLifetime: time.Duration(config.MessageLifetimeMinutes) * time.Minute,
	}
}

func (u *ChatUsecase) SendMessage(ctx context.Context, msg *entities.ChatMessage) error {
	if len(msg.Content) > u.maxMessageLen {
		return fmt.Errorf("сообщение слишком длинное (максимум %d символов)", u.maxMessageLen)
	}
	if msg.Content == "" {
		return errors.ErrEmptyMessage
	}

	u.logger.Info("отправка сообщения в чат",
		logger.NewField("user_id", msg.UserID),
		logger.NewField("content_len", len(msg.Content)),
	)
	return u.repo.SaveMessage(ctx, msg.UserID, msg.Username, msg.Content)
}

func (u *ChatUsecase) GetMessages(ctx context.Context) ([]*entities.ChatMessage, error) {
	return u.repo.GetMessages(ctx)
}

func (u *ChatUsecase) DeleteOldMessages(ctx context.Context, before time.Time) error {
	u.logger.Info("deleting old messages",
		logger.NewField("before", before))
	return u.repo.DeleteOldMessages(ctx, before)
}

type CleanupService struct {
	chatUC  ChatUsecaseInterface
	logger  logger.Logger
	ticker  *time.Ticker
	done    chan bool
	timeout time.Duration
}

func NewCleanupService(chatUC ChatUsecaseInterface, logger logger.Logger) *CleanupService {
	return &CleanupService{
		chatUC:  chatUC,
		logger:  logger,
		done:    make(chan bool),
		timeout: 30 * time.Second,
	}
}

func (s *CleanupService) Start(interval time.Duration, messageLifetime time.Duration) {
	s.ticker = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-s.ticker.C:
				if err := s.Cleanup(messageLifetime); err != nil {
					s.logger.Error("failed to cleanup messages",
						logger.NewField("error", err))
				}
			case <-s.done:
				s.ticker.Stop()
				return
			}
		}
	}()
}

func (s *CleanupService) Stop() {
	s.done <- true
}

func (s *CleanupService) Cleanup(messageLifetime time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	cutoff := time.Now().Add(-messageLifetime)

	if err := s.chatUC.DeleteOldMessages(ctx, cutoff); err != nil {
		s.logger.Error("ошибка очистки старых сообщений",
			logger.NewField("error", err),
			logger.NewField("cutoff", cutoff),
		)
		return err
	}

	s.logger.Info("успешная очистка старых сообщений",
		logger.NewField("cutoff", cutoff),
	)
	return nil
}

type PostUsecaseInterface interface {
	CreatePost(ctx context.Context, post *entities.Post) error
	GetPostByID(ctx context.Context, id int64) (*entities.Post, error)
	UpdatePost(ctx context.Context, post *entities.Post) error
	DeletePost(ctx context.Context, id int64) error
	Posts(ctx context.Context) ([]*entities.Post, error)
}

type PostUsecase struct {
	repo   repository.PostRepository
	logger logger.Logger
}

func NewPostUsecase(repo repository.PostRepository, logger logger.Logger) *PostUsecase {
	return &PostUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (u *PostUsecase) CreatePost(ctx context.Context, post *entities.Post) error {
	u.logger.Info("создание нового поста",
		logger.NewField("title", post.Title),
		logger.NewField("author_id", post.AuthorID))

	return u.repo.CreatePost(ctx, post)
}

func (u *PostUsecase) GetPostByID(ctx context.Context, id int64) (*entities.Post, error) {
	u.logger.Info("получение поста по ID",
		logger.NewField("post_id", id))

	return u.repo.GetPostByID(ctx, id)
}

func (u *PostUsecase) UpdatePost(ctx context.Context, post *entities.Post) error {
	u.logger.Info("обновление поста",
		logger.NewField("post_id", post.ID))
	return u.repo.UpdatePost(ctx, post)
}

func (u *PostUsecase) DeletePost(ctx context.Context, id int64) error {
	u.logger.Info("удаление поста по ID",
		logger.NewField("post_id", id))
	return u.repo.DeletePost(ctx, id)
}

func (u *PostUsecase) Posts(ctx context.Context) ([]*entities.Post, error) {
	return u.repo.Posts(ctx)
}

type CommentUsecaseInterface interface {
	CreateComment(ctx context.Context, comment *entities.Comment) error
	GetCommentByID(ctx context.Context, id int64) (*entities.Comment, error)
	UpdateComment(ctx context.Context, comment *entities.Comment) error
	DeleteComment(ctx context.Context, id int64) error
	GetByPostID(ctx context.Context, postID int64) ([]*entities.Comment, error)
	GetByUserID(ctx context.Context, userID int64) ([]*entities.Comment, error)
}

type CommentUsecase struct {
	repo   repository.CommentRepository
	logger logger.Logger
}

func NewCommentUsecase(repo repository.CommentRepository, logger logger.Logger) *CommentUsecase {
	return &CommentUsecase{
		repo:   repo,
		logger: logger,
	}
}

func (u *CommentUsecase) CreateComment(ctx context.Context, comment *entities.Comment) error {
	u.logger.Info("создание нового комментария",
		logger.NewField("comment_id", comment.ID))
	return u.repo.CreateComment(ctx, comment)
}

func (u *CommentUsecase) GetCommentByID(ctx context.Context, id int64) (*entities.Comment, error) {
	u.logger.Info("получение комментария по ID",
		logger.NewField("comment_id", id))
	return u.repo.GetCommentByID(ctx, id)
}

func (u *CommentUsecase) GetByPostID(ctx context.Context, postID int64) ([]*entities.Comment, error) {
	u.logger.Info("получение комментариев по ID поста",
		logger.NewField("post_id", postID))
	return u.repo.GetByPostID(ctx, postID)
}

func (u *CommentUsecase) GetByUserID(ctx context.Context, userID int64) ([]*entities.Comment, error) {
	u.logger.Info("получение комментариев по ID пользователя",
		logger.NewField("user_id", userID))
	return u.repo.GetByUserID(ctx, userID)
}

func (u *CommentUsecase) UpdateComment(ctx context.Context, comment *entities.Comment) error {
	u.logger.Info("обновление комментария",
		logger.NewField("comment_id", comment.ID))
	return u.repo.UpdateComment(ctx, comment)
}

func (u *CommentUsecase) DeleteComment(ctx context.Context, commentId int64) error {
	u.logger.Info("удаление комментария",
		logger.NewField("comment_id", commentId))

	u.logger.Info("YA TUT")
	return u.repo.DeleteComment(ctx, commentId)
}
