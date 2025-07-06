package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"

	pb "github.com/netabakovv/forum/back/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func dialForumService(t *testing.T) pb.ForumServiceClient {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(5*time.Second))
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })

	return pb.NewForumServiceClient(conn)
}

func TestForumService(t *testing.T) {
	client := dialForumService(t)
	ctx := context.Background()

	t.Run("CreatePost_Positive", func(t *testing.T) {
		resp, err := client.CreatePost(ctx, &pb.CreatePostRequest{
			Title:          "Test post",
			Content:        "Test content",
			AuthorId:       1,
			AuthorUsername: "tester",
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotZero(t, resp.Post.Id)
		require.Equal(t, "Test post", resp.Post.Title)
	})

	t.Run("CreatePost_Negative_EmptyTitle", func(t *testing.T) {
		_, err := client.CreatePost(ctx, &pb.CreatePostRequest{
			Title:          "",
			Content:        "Content",
			AuthorId:       1,
			AuthorUsername: "tester",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	var createdPostID int64

	t.Run("CreatePost_SaveIDForFurther", func(t *testing.T) {
		resp, err := client.CreatePost(ctx, &pb.CreatePostRequest{
			Title:          "Post to update",
			Content:        "Original content",
			AuthorId:       2,
			AuthorUsername: "author2",
		})
		require.NoError(t, err)
		require.NotZero(t, resp.Post.Id)
		createdPostID = resp.Post.Id
	})

	t.Run("GetPost_Positive", func(t *testing.T) {
		resp, err := client.GetPost(ctx, &pb.GetPostRequest{
			PostId: createdPostID,
		})
		require.NoError(t, err)
		require.Equal(t, createdPostID, resp.Post.Id)
	})

	t.Run("UpdatePost_Positive", func(t *testing.T) {
		newTitle := "Updated title"
		newContent := "Updated content"
		resp, err := client.UpdatePost(ctx, &pb.UpdatePostRequest{
			PostId:  createdPostID,
			Title:   &newTitle,
			Content: &newContent,
		})
		require.NoError(t, err)
		require.Equal(t, newTitle, resp.Post.Title)
		require.Equal(t, newContent, resp.Post.Content)
	})

	t.Run("DeletePost_Positive", func(t *testing.T) {
		_, err := client.DeletePost(ctx, &pb.DeletePostRequest{
			PostId: createdPostID,
		})
		require.NoError(t, err)
	})

	t.Run("Posts_Positive", func(t *testing.T) {
		resp, err := client.Posts(ctx, &pb.ListPostsRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		// Просто проверяем, что получили слайс постов (даже если пустой)
		require.NotNil(t, resp.Posts)
	})

	// ============================
	// Комментарии
	// ============================

	var commentID int64
	t.Run("CreateComment_Positive", func(t *testing.T) {
		resp, err := client.CreateComment(ctx, &pb.CreateCommentRequest{
			Content:        "Test comment",
			AuthorId:       1,
			PostId:         1, // Подразумевается, что пост с ID=1 есть
			AuthorUsername: "tester",
		})
		require.NoError(t, err)
		require.NotZero(t, resp.Comment.Id)
		commentID = resp.Comment.Id
	})

	t.Run("CreateComment_Negative_EmptyContent", func(t *testing.T) {
		_, err := client.CreateComment(ctx, &pb.CreateCommentRequest{
			Content:        "",
			AuthorId:       1,
			PostId:         1,
			AuthorUsername: "tester",
		})
		require.Error(t, err)
		st, _ := status.FromError(err)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("GetCommentByID_Positive", func(t *testing.T) {
		resp, err := client.GetCommentByID(ctx, &pb.GetCommentRequest{
			CommentId: commentID,
		})
		require.NoError(t, err)
		require.Equal(t, commentID, resp.Comment.Id)
	})

	t.Run("GetCommentByID_Negative_ZeroID", func(t *testing.T) {
		_, err := client.GetCommentByID(ctx, &pb.GetCommentRequest{CommentId: 0})
		require.Error(t, err)
		st, _ := status.FromError(err)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("GetByPostID_Positive", func(t *testing.T) {
		resp, err := client.GetByPostID(ctx, &pb.GetCommentsByPostIDRequest{PostId: 1})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})

	t.Run("UpdateComment_Positive", func(t *testing.T) {
		newContent := "Updated comment content"
		resp, err := client.UpdateComment(ctx, &pb.UpdateCommentRequest{
			CommentId: commentID,
			Content:   &newContent,
		})
		require.NoError(t, err)
		require.Equal(t, newContent, resp.Comment.Content)
	})

	t.Run("UpdateComment_Negative_ZeroID", func(t *testing.T) {
		text := "text"
		_, err := client.UpdateComment(ctx, &pb.UpdateCommentRequest{CommentId: 0, Content: &text})
		require.Error(t, err)
		st, _ := status.FromError(err)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	// ============================
	// Чат
	// ============================

	t.Run("SendMessage_Positive", func(t *testing.T) {
		_, err := client.SendMessage(ctx, &pb.ChatMessage{
			UserId:  1,
			Content: "Hello from test",
		})
		require.NoError(t, err)
	})

	t.Run("SendMessage_Negative_EmptyContent", func(t *testing.T) {
		_, err := client.SendMessage(ctx, &pb.ChatMessage{
			UserId:  1,
			Content: "",
		})
		require.Error(t, err)
		st, _ := status.FromError(err)
		require.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("GetMessages_Positive", func(t *testing.T) {
		resp, err := client.GetMessages(ctx, &pb.GetMessagesRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
	})
}

const baseURL = "http://localhost:8090"

var accessToken string
var createdPostID int64
var createdCommentID int64

func authHeader() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + accessToken,
	}
}

func doRequest(t *testing.T, method, url string, body any, headers map[string]string) *http.Response {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err)

	if headers != nil {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func TestFullFlow(t *testing.T) {
	t.Run("Register", func(t *testing.T) {
		resp := doRequest(t, "POST", baseURL+"/register", map[string]string{
			"username": "testuser",
			"password": "testpass",
		}, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Login", func(t *testing.T) {
		resp := doRequest(t, "POST", baseURL+"/login", map[string]string{
			"username": "testuser",
			"password": "testpass",
		}, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body struct {
			AccessToken string `json:"access_token"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
		require.NotEmpty(t, body.AccessToken)

		accessToken = body.AccessToken
	})

	t.Run("GetProfile", func(t *testing.T) {
		resp := doRequest(t, "GET", baseURL+"/api/profile", nil, authHeader())
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("CreatePost", func(t *testing.T) {
		resp := doRequest(t, "POST", baseURL+"/api/posts", map[string]string{
			"title":   "Test Post",
			"content": "Test Content",
		}, authHeader())
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetPosts", func(t *testing.T) {
		resp := doRequest(t, "GET", baseURL+"/posts", nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var posts []map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&posts))
		require.NotEmpty(t, posts)

		id := posts[0]["id"]
		switch id := id.(type) {
		case float64:
			createdPostID = int64(id)
		case int64:
			createdPostID = id
		}
	})

	t.Run("GetCommentsByPostID", func(t *testing.T) {
		resp := doRequest(t, "GET", baseURL+"/comments/post/"+strconv.FormatInt(1, 10), nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var comments []map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&comments))
		require.NotEmpty(t, comments)

		id := comments[0]["id"]
		switch id := id.(type) {
		case float64:
			createdCommentID = int64(id)
		case int64:
			createdCommentID = id
		}
	})

	t.Run("GetMessages", func(t *testing.T) {
		resp := doRequest(t, "GET", baseURL+"/chat", nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var messages []map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&messages))
		require.NotEmpty(t, messages)
	})

	t.Run("DeletePost", func(t *testing.T) {
		resp := doRequest(t, "DELETE", baseURL+"/api/posts/"+strconv.FormatInt(createdPostID, 10), nil, authHeader())
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Logout", func(t *testing.T) {
		body := map[string]string{
			"access_token": accessToken,
		}
		resp := doRequest(t, "POST", baseURL+"/api/logout", body, authHeader())
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

}
