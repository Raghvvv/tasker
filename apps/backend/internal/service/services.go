package service

import (
	"fmt"

	"github.com/Raghvvv/tasker/internal/lib/aws"
	"github.com/Raghvvv/tasker/internal/lib/job"
	"github.com/Raghvvv/tasker/internal/repository"
	"github.com/Raghvvv/tasker/internal/server"
)

type Services struct {
	Auth       *AuthService
	Job        *job.JobService
	Todo       *TodoService
	Comment    *CommentService
	Categories *CategoryService
	// awsClient	*aws.AWS
}

func NewServices(s *server.Server, repos *repository.Repositories) (*Services, error) {
	authService := NewAuthService(s)
	awsClient, error := aws.NewAWS(s)
	if error != nil {
		return nil, fmt.Errorf("failed to createt aws client%w", error)
	}
	return &Services{
		Job:        s.Job,
		Auth:       authService,
		Comment:    NewCommentService(s, repos.Comment, repos.Todo),
		Categories: NewCategoryService(s, repos.Category),
		Todo:       CreateTodoService(s, repos.Todo, repos.Category, awsClient),
	}, nil
}
