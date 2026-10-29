package category

import "github.com/Raghvvv/tasker/internal/model"

type Category struct {
	model.Base
	UserID      string  `json:"userID" db:"user_id"`
	Name        string  `json:"name" db:"name"`
	Colour      string  `json:"colour" db:"colour"`
	Description *string `json:"description" db:"description"`
}
