package handler

import (
	"github.com/Raghvvv/tasker/internal/server"
	"github.com/Raghvvv/tasker/internal/service"
)

type Handlers struct {
	Health   *HealthHandler
	OpenAPI  *OpenAPIHandler
	Todo     *TodoHandler
	Comment  *CommentHandler
	Category *CategoryHandler
}

func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:   NewHealthHandler(s),
		OpenAPI:  NewOpenAPIHandler(s),
		Todo:     NewTodoHandler(s, services.Todo),
		Comment:  NewCommentHandler(s, services.Comment),
		Category: NewCategoryHandler(s, services.Categories),
	}
}
