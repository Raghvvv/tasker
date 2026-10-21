package comment

import (
	"github.com/Raghvvv/tasker/internal/model"
	"github.com/google/uuid"
)

type comment struct {
	model.Base
	TodoId  uuid.UUID `json:"todoId" db:"todo_id"`
	UserID  string    `json:"userId" db:"user_id"`
	Content string    `json:"content" db:"content"`
}
