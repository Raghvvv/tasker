package handler

import (
	// "go/doc/comment"
	"net/http"

	"github.com/Raghvvv/tasker/internal/middleware"
	"github.com/Raghvvv/tasker/internal/model/comment"
	"github.com/Raghvvv/tasker/internal/server"
	"github.com/Raghvvv/tasker/internal/service"
	"github.com/labstack/echo/v4"
)

type CommentHandler struct {
	Handler
	commentService *service.CommentService
}

func NewCommentHandler(s *server.Server, commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{
		Handler:        NewHandler(s),
		commentService: commentService,
	}
}
func (h CommentHandler) CreateComment(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *comment.CreateCommentPayload) (*comment.Comment, error) {
			userId := middleware.GetRequestID(c)
			return h.commentService.AddComment(c, userId, payload.TodoId, payload)
		},
		http.StatusCreated,
		&comment.CreateCommentPayload{},
	)(c)

}
func (h CommentHandler) GetCommentByID(c echo.Context) error {

	return Handle(
		h.Handler,
		func(c echo.Context, payload *comment.FetchCommentByTodoIdPayload) ([]comment.Comment, error) {
			userId := middleware.GetUserID(c)
			return h.commentService.GetCommentsByTodoID(c, userId, payload.TodoId)
		},
		http.StatusOK,
		&comment.FetchCommentByTodoIdPayload{},
	)(c)
}
func (h CommentHandler) UpdateComment(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *comment.UpdateCommentPayload) (*comment.Comment, error) {
			userId := middleware.GetUserID(c)
			return h.commentService.UpdateComment(c, userId, payload.ID, payload.Content)
		},
		http.StatusOK,
		&comment.UpdateCommentPayload{},
	)(c)

}

func (h *CommentHandler) DeleteComment(c echo.Context) error {
	return HandleNoContent(
		h.Handler,
		func(c echo.Context, payload *comment.DeleteCommentPayload) error {
			userID := middleware.GetUserID(c)
			return h.commentService.DeleteComment(c, userID, payload.ID)
		},
		http.StatusNoContent,
		&comment.DeleteCommentPayload{},
	)(c)
}
