package v1

import (
	"github.com/Raghvvv/tasker/internal/handler"
	"github.com/Raghvvv/tasker/internal/middleware"
	"github.com/labstack/echo/v4"
)

func registerTodoRoutes(r *echo.Group, h *handler.TodoHandler, ch *handler.CommentHandler, auth *middleware.AuthMiddleware) {
	todos := r.Group("/todos")

	todos.Use(auth.RequireAuth)
	todos.POST("", h.CreateTodo)
	// todos.GET("/stats",h.GetTodoStats)
	todos.GET("", h.GetTodos)

	dynamicTodos := todos.Group("/:id")
	dynamicTodos.GET("", h.GetTodoByID)
	dynamicTodos.PATCH("", h.UpdateTodo)
	dynamicTodos.DELETE("", h.DeleteTodo)

	todoComments := dynamicTodos.Group("/comments")
	todoComments.POST("", ch.CreateComment)
	todoComments.GET("", ch.GetCommentByID)

}
