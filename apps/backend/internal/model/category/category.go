package category

import "github.com/Raghvvv/tasker/internal/model"

type Category struct {
	model.Base
	UserID      string  `json:"userID" db:"user_id"`
	Name        string  `json:"name" db:"name"`
	Color       string  `json:"color" db:"color"`
	Description *string `json:"description" db:"description"`
}
