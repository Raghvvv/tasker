package service

import (
	// "uuid"

	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/Raghvvv/tasker/internal/errs"
	"github.com/Raghvvv/tasker/internal/lib/aws"

	// "github.com/Raghvvv/tasker/internal/logger"
	"github.com/Raghvvv/tasker/internal/middleware"
	"github.com/Raghvvv/tasker/internal/model"
	"github.com/Raghvvv/tasker/internal/model/todo"
	"github.com/Raghvvv/tasker/internal/repository"
	"github.com/Raghvvv/tasker/internal/server"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type TodoService struct {
	server       *server.Server
	todoRepo     *repository.TodoRepository
	categoryRepo *repository.CategoryRepository
	awsClient    *aws.AWS
}

func CreateTodoService(s *server.Server, todoRepo *repository.TodoRepository, categoryRepo *repository.CategoryRepository, awsClient *aws.AWS) *TodoService {
	return &TodoService{
		server:       s,
		todoRepo:     todoRepo,
		categoryRepo: categoryRepo,
		awsClient:    awsClient,
	}
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

//		return stats, nil
//	}
func (s *TodoService) UploadTodoAttatchment(ctx echo.Context, userID string, todoId uuid.UUID, file *multipart.FileHeader) (*todo.TodoAttatchment, error) {
	logger := middleware.GetLogger(ctx)
	// todoUUID,error:=uuid.Parse(todoId)
	_, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoId)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed")
		return nil, err
	}
	//open file:
	src, err := file.Open()
	if err != nil {
		logger.Error().Err(err).Msg("failed to open file")
		return nil, errs.NewBadRequestError("failed to open file", false, nil, nil, nil)
	}
	defer src.Close()

	//upload to s3:
	key, err := s.awsClient.S3.UploadFiles(
		ctx.Request().Context(),
		s.server.Config.AWS.UploadBucket,
		"todos/attatchments"+file.Filename,
		src,
	)

	if err != nil {
		logger.Error().Err(err).Msg("failed to upload file to S3")
		return nil, errors.Wrap(err, "failed to upload file")
	}

	//detect mime:
	src, err = file.Open()
	if err != nil {
		logger.Error().Err(err).Msg("failed to reopen file for mime detection")
		return nil, errs.NewBadRequestError("failed to open file", false, nil, nil, nil)
	}
	defer src.Close()

	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		logger.Error().Err(err).Msg("failed to read file for mime detection")
		return nil, errs.NewBadRequestError("failed to process file", false, nil, nil, nil)
	}
	mimeType := http.DetectContentType(buffer)

	// Create attachment record
	attachment, err := s.todoRepo.UploadTodoAttatchment(
		ctx.Request().Context(),
		todoId,
		userID,
		key,
		file.Filename,
		file.Size,
		mimeType,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create attachment record")
		return nil, err
	}

	logger.Info().
		Str("attachment_id", attachment.ID.String()).
		Str("s3_key", key).
		Msg("uploaded todo attachment")

	return attachment, nil

}

func (s *TodoService) DeleteTodoAttatchment(todoID uuid.UUID, ctx echo.Context, userID string, attchmentID uuid.UUID) error {
	logger := middleware.GetLogger(ctx)
	//check todo exists:
	_, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo doesnt exist!")
		return errs.NewBadRequestError("todo doesnt exist!", false, nil, nil, nil)

	}
	//get attatchment details:
	attatchment, err := s.todoRepo.GetTodoAttatchment(ctx.Request().Context(), todoID, attchmentID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get attatchment details")
		return err
	}
	//delete from DB:
	err = s.todoRepo.DeleteTodoAttatchment(
		ctx.Request().Context(),
		todoID,
		attchmentID,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to delete attatchment from DB")
		return err
	}

	//delete from s3:
	go func() {
		err = s.awsClient.S3.DeleteAttatchment(
			ctx.Request().Context(),
			s.server.Config.AWS.UploadBucket,
			attatchment.DownloadKey,
		)
		if err != nil {
			s.server.Logger.Error().
				Err(err).
				Str("s3_key", attatchment.DownloadKey).
				Msg("failed to delete attachment from S3")
		}
	}()

	logger.Info().Msg("deleted todo attachment")

	return nil

}

func (s *TodoService) GetAttatchmentPresignedURL(
	ctx echo.Context,
	userID string,
	todoID uuid.UUID,
	attachmentID uuid.UUID,
) (string, error) {
	logger := middleware.GetLogger(ctx)

	// Verify todo exists and belongs to user
	_, err := s.todoRepo.CheckTodoExists(ctx.Request().Context(), userID, todoID)
	if err != nil {
		logger.Error().Err(err).Msg("todo validation failed")
		return "", err
	}

	// Get attachment details
	attachment, err := s.todoRepo.GetTodoAttatchment(
		ctx.Request().Context(),
		todoID,
		attachmentID,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get attachment details")
		return "", err
	}

	// Generate presigned URL
	url, err := s.awsClient.S3.CreatePresignedURL(
		ctx.Request().Context(),
		s.server.Config.AWS.UploadBucket,
		attachment.DownloadKey,
	)
	if err != nil {
		logger.Error().Err(err).Msg("failed to generate presigned URL")
		return "", err
	}
	fmt.Println(url)
	return url, nil
}
