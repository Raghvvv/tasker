package service

import (
	// "uuid"

	"github.com/Raghvvv/tasker/internal/errs"
	"github.com/Raghvvv/tasker/internal/middleware"
	"github.com/Raghvvv/tasker/internal/model"
	"github.com/Raghvvv/tasker/internal/model/todo"
	"github.com/Raghvvv/tasker/internal/repository"
	"github.com/Raghvvv/tasker/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TodoService struct {
	server       *server.Server
	todoRepo     *repository.TodoRepository
	categoryRepo *repository.CategoryRepository
}

func CreateTodoService(s *server.Server, todoRepo *repository.TodoRepository, categoryRepo *repository.CategoryRepository) (*TodoService, error) {
	return &TodoService{
		server:       s,
		todoRepo:     todoRepo,
		categoryRepo: categoryRepo,
	}, nil
}

func (s *TodoService) CreateTodo(ctx echo.Context, payload *todo.CreateTodoPayload, userID string) (*todo.Todo, error) {
	logger := middleware.GetLogger(ctx)

	if payload.ParentTodoID != nil {
		parentTodo, error := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, *payload.ParentTodoID)
		if error != nil {
			logger.Error().Err(error).Msg("parent todo validation failed")
			return nil, error
		}
		if !parentTodo.CanHaveChildren() {
			err := errs.NewBadRequestError("subtodo cannot have child todo", false, nil, nil, nil)
			logger.Warn().Msg("parentTodo cant have childern")
			return nil, err
		}
	}
	//validate categoryID exists and belongs to the user (if provided)
	if payload.CategoryID != nil {
		_, error := s.categoryRepo.GetCategoryByID(ctx.Request().Context(), userID, *payload.CategoryID)
		if error != nil {
			logger.Error().Err(error).Msg("category validation failed")
			return nil, error
		}

	}
	todo, error := s.todoRepo.CreateTodo(ctx.Request().Context(), userID, payload)
	if error != nil {
		logger.Error().Err(error).Msg("failed to create todo")
		return nil, error
	}
	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo_created").
		Str("todo_id", todo.ID.String()).
		Str("title", todo.Title).
		Str("category_id", func() string {
			if todo.CategoryID != nil {
				return todo.CategoryID.String()
			}
			return ""
		}()).
		Str("priority", string(*todo.Priority)).
		Msg("Todo created successfully")

	return todo, nil
}

func (s *TodoService) GetTodos(ctx echo.Context, userID string, query *todo.GetTodoQuery) (*model.PaginatedResponse[todo.Populatedtodo], error) {
	logger := middleware.GetLogger(ctx)
	result, err := s.todoRepo.GetTodos(ctx.Request().Context(), userID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch todos")
	}
	return result, nil
}

func (s *TodoService) GetTodoByID(ctx echo.Context, userID string, todoID uuid.UUID) (*todo.Populatedtodo, error) {
	logger := middleware.GetLogger(ctx)
	todoItem, err := s.todoRepo.GetTodoByID(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch todo")
		return nil, err
	}
	return todoItem, nil
}

func (s *TodoService) UpdateTodo(ctx echo.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
	logger := middleware.GetLogger(ctx)
	//validate todo
	if payload.ParentTodoID != nil {
		parentTodo, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, *payload.ParentTodoID)
		if err != nil {
			logger.Error().Err(err).Msg("parent validation failed")
			return nil, err
		}
		if parentTodo.ID == payload.ID {
			err := errs.NewBadRequestError("todo cant be its own parent", false, nil, nil, nil)
			logger.Warn().Msg("todo cant be itsw own parent")
			return nil, err
		}
		if !parentTodo.CanHaveChildren() {
			err := errs.NewBadRequestError("subtodos cant have subtodos", false, nil, nil, nil)
			logger.Warn().Msg("subtodos cant have subtodos")
			return nil, err
		}
		logger.Debug().Msg("parent validation passed")
	}
	//validate category:
	if payload.CategoryID != nil {
		_, err := s.categoryRepo.GetCategoryByID(ctx.Request().Context(), userID, *payload.CategoryID)
		if err != nil {
			logger.Error().Err(err).Msg("category validation failed")
			return nil, err
		}
		logger.Debug().Msg("category validation passed")
	}

	updatedTodo, err := s.todoRepo.UpdateTodo(ctx.Request().Context(), userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to update todo")
		return nil, err
	}
	//business event log:
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo updated").
		Str("todo_id", updatedTodo.ID.String()).
		Str("title", updatedTodo.Title).
		Str("category_id", func() string {
			if updatedTodo.CategoryID != nil {
				return string(updatedTodo.CategoryID.String())
			}
			return ""
		}()).
		Str("priority", string(*updatedTodo.Priority)).
		Str("status", string(updatedTodo.Status)).
		Msg("todo updated successfully")

	return updatedTodo, nil

}
func (s *TodoService) DeleteTodo(ctx echo.Context, userID string, todoID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)

	err := s.todoRepo.DeleteTodo(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete todo")
		return err
	}

	// Business event log
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "todo_deleted").
		Str("todo_id", todoID.String()).
		Msg("Todo deleted successfully")

	return nil
}

// func (s *TodoService) GetTodoStats(ctx echo.Context, userID string) (*todo.TodoStats, error) {
// 	logger := middleware.GetLogger(ctx)

// 	stats, err := s.todoRepo.GetTodoStats(ctx.Request().Context(), userID)
// 	if err != nil {
// 		logger.Error().Err(err).Msg("failed to fetch todo statistics")
// 		return nil, err
// 	}

// 	return stats, nil
// }
