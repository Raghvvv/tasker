package todo

import (
	"github.com/Raghvvv/tasker/internal/model"
	"github.com/google/uuid"
)

type TodoAttatchment struct {
	model.Base
	TodoID      uuid.UUID `json:"todoID" db:"todo_id"`
	FileName    string    `json:"fileName" db:"file_name"`
	UploadedBy  string    `json:"uploadedBy" db:"uploaded_by"`
	DownloadKey string    `json:"downloadKey" db:"download_key"`
	FileSize    *int64    `json:"fileSize" db:"file_size"`
	MimeType    *string   `json:"mimeType" db:"mime_type"`
}
