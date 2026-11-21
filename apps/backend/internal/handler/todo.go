package handler

import (
	"net/http"

	"github.com/Raghvvv/tasker/internal/middleware"
	"github.com/Raghvvv/tasker/internal/model"

	"github.com/Raghvvv/tasker/internal/model/todo"
	"github.com/Raghvvv/tasker/internal/server"
	"github.com/Raghvvv/tasker/internal/service"
	"github.com/labstack/echo/v4"
)

type TodoHandler struct {
	Handler
	todoservice *service.TodoService
}

func NewTodoHandler(s *server.Server, todoservice *service.TodoService) *TodoHandler {
	return &TodoHandler{
		Handler:     NewHandler(s),
		todoservice: todoservice,
	}
}

func (h *TodoHandler) CreateTodo(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
			userID := middleware.GetUserID(c)
			return h.todoservice.CreateTodo(c, payload, userID)
		},
		http.StatusCreated,
		&todo.CreateTodoPayload{},
	)(c)
}

func (h *TodoHandler) GetTodoByID(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.GetTodoByIDPayload) (*todo.Populatedtodo, error) {
			userID := middleware.GetUserID(c)
			return h.todoservice.GetTodoByID(c, userID, payload.ID)
		},
		http.StatusOK,
		&todo.GetTodoByIDPayload{},
	)(c)
}
func (h *TodoHandler) GetTodos(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.GetTodoQuery) (*model.PaginatedResponse[todo.Populatedtodo], error) {
			userID := middleware.GetUserID(c)
			return h.todoservice.GetTodos(c, userID, payload)
		},
		http.StatusOK,
		&todo.GetTodoQuery{},
	)(c)
}
func (h *TodoHandler) UpdateTodo(c echo.Context) error {
	return Handle(
		h.Handler,
		func(c echo.Context, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
			userID := middleware.GetUserID(c)
			return h.todoservice.UpdateTodo(c, userID, payload)

		},
		http.StatusOK,
		&todo.UpdateTodoPayload{},
	)(c)

}
func (h *TodoHandler) DeleteTodo(c echo.Context) error {
	return HandleNoContent(
		h.Handler,
		func(c echo.Context, payload *todo.DeleteTodoPayload) error {
			userID := middleware.GetUserID(c)
			return h.todoservice.DeleteTodo(c, userID, payload.ID)
		},
		http.StatusNoContent,
		&todo.DeleteTodoPayload{},
	)(c)

}

// func (h *TodoHandler) GetTodoStats(c echo.Context) error {
// 	return Handle(
// 		h.Handler,
// 		func(c echo.Context, payload *todo.GetTodoStatsPayload) (*todo.TodoStats, error) {
// 			userID := middleware.GetUserID(c)
// 			return h.todoservice.GetTodoStats(c, userID)
// 		},
// 		http.StatusOK,
// 		&todo.GetTodoStatsPayload{},
// 	)(c)
// }
