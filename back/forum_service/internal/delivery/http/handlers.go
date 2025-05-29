package http

import (
	"back/forum_service/internal/entities"
	"back/forum_service/internal/usecase"
	"back/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HandlerInterface interface {
	SendMessage(c *gin.Context)
	GetMessages(c *gin.Context)
}

type Handler struct {
	chatUC    usecase.ChatUsecaseInterface
	postUC    usecase.PostUsecase
	commentUC usecase.CommentUsecaseInterface
	log       logger.Logger
}

func NewHandler(chatUC usecase.ChatUsecaseInterface, postUC usecase.PostUsecase, commentUC usecase.CommentUsecaseInterface, log logger.Logger) *Handler {
	return &Handler{
		chatUC:    chatUC,
		postUC:    postUC,
		commentUC: commentUC,
		log:       log,
	}
}

func (h *Handler) SendMessage(c *gin.Context) {
	var msg entities.ChatMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.chatUC.SendMessage(c.Request.Context(), &msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) GetMessages(c *gin.Context) {
	messages, err := h.chatUC.GetMessages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *Handler) CreatePost(c *gin.Context) {
	var post entities.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.postUC.CreatePost(c.Request.Context(), &post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) GetPostByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	post, err := h.postUC.GetPostByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *Handler) GetAllPosts(c *gin.Context) {
	posts, err := h.postUC.Posts(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, posts)
}

func (h *Handler) UpdatePost(c *gin.Context) {
	var post entities.Post
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	post.ID = id

	if err := h.postUC.UpdatePost(c.Request.Context(), &post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) DeletePost(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.postUC.DeletePost(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) CreateComment(c *gin.Context) {
	var comment entities.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.commentUC.CreateComment(c.Request.Context(), &comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *Handler) GetCommentByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	comment, err := h.commentUC.GetCommentByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comment)
}

func (h *Handler) GetCommentsByPostID(c *gin.Context) {
	postID, _ := strconv.ParseInt(c.Param("post_id"), 10, 64)
	comments, err := h.commentUC.GetByPostID(c.Request.Context(), postID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comments)
}

func (h *Handler) GetCommentsByUserID(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.Param("user_id"), 10, 64)
	comments, err := h.commentUC.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comments)
}

func (h *Handler) UpdateComment(c *gin.Context) {
	var comment entities.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	comment.ID = id

	if err := h.commentUC.UpdateComment(c.Request.Context(), &comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) DeleteComment(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.commentUC.DeleteComment(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
