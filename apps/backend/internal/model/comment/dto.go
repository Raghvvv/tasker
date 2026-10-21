package comment

import (
	// "uuid"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type FetchCommentByTodoIdPayload struct {
	TodoId uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *FetchCommentByTodoIdPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type CreateCommentPayload struct {
	TodoId  uuid.UUID `param:"id" validate:"required,uuid"`
	Content string    `json:"content" Validate:"required,min=1,max=100"`
}

func (p *CreateCommentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type DeleteCommentPayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *DeleteCommentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type UpdateCommentPayload struct {
	ID      uuid.UUID `param:"id" validate:"required,uuid"`
	Content string    `json:"content" validate:"required, min=1,max=1000"`
}

func (p *UpdateCommentPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}
