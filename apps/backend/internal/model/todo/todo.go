package todo

import (
	"time"
	// "uuid"

	"github.com/Raghvvv/tasker/internal/model"
	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusCompleted Status = "completed"
	StatusActive    Status = "active"
	StatusArchieved Status = "archieved"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
)

type Metadata struct {
	Tags       []string `json:"tags" db:"tags"`
	Colour     *string  `json:"colour" db:"colour"`
	Difficulty *int     `json:"difficulty" db:"difficulty"`
	Reminder   *string  `json:"reminder" db:"reminder"`
}

type Todo struct {
	model.Base

	UserID       string     `json:"userId" db:"user_id"`
	Title        string     `json:"title" db:"title"`
	Description  *string    `json:"description" db:"description"`
	Status       Status     `json:"status" db:"status"`
	Priority     *Priority  `json:"priority" db:"priority"`
	DueDate      *time.Time `json:"dueDate" db:"due_date"`
	CompletedAt  *time.Time `json:"completedAt" db:"completed_at"`
	ParentTodoID *uuid.UUID `json:"parentTodoId" db:"parent_todo_id"`
	CategoryID   *uuid.UUID `json:"categoryId" db:"category_id"`
	Metadata     *Metadata  `json:"metadata" db:"metadata"`
	SortOrder    int        `json:"sortOrder" db:"sort_order"`
}
