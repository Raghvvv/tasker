package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Raghvvv/tasker/internal/model"
	"github.com/Raghvvv/tasker/internal/model/category"
	"github.com/Raghvvv/tasker/internal/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CategoryRepository struct {
	server *server.Server
}

func NewCategoryRepository(s *server.Server) *CategoryRepository {
	return &CategoryRepository{server: s}
}

func (r *CategoryRepository) AddCategory(ctx context.Context, userID string, payload *category.CreateCategoryPayload) (*category.Category, error) {
	stmt := `INSERT INTO 
			todo_categories	
			(
				user_id,
				name,
				color,
				description
			) 
			VALUES(
				@user_id,
				@name,
				@color,
				@decription
			)
			RETURNING *
			`
	args := pgx.NamedArgs{
		"user_id":     userID,
		"name":        payload.Name,
		"colour":      payload.Name,
		"description": *payload.Description,
	}
	rows, error := r.server.DB.Pool.Query(ctx, stmt, args)
	if error != nil {
		return nil, fmt.Errorf("failed to execute Query and get rows userID=%s:%w", userID, error)
	}
	todoCategoryItem, error := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if error != nil {
		return nil, fmt.Errorf("failed to populate category struct from rows userID=%s:%w ", userID, error)
	}
	return &todoCategoryItem, nil
}

func (r *CategoryRepository) GetCategoryByID(ctx context.Context, userID string, categoryID uuid.UUID) (*category.Category, error) {
	stmt := `SELECT * FROM todo_categories WHERE user_id=@userID AND id=@categoryID`
	args := pgx.NamedArgs{
		"userID": userID,
		"id":     categoryID,
	}
	rows, error := r.server.DB.Pool.Query(ctx, stmt, args)
	if error != nil {
		return nil, fmt.Errorf("failed to execute Query for user=%s:%w", userID, error)
	}
	category, error := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if error != nil {
		return nil, fmt.Errorf("failed to populate struct from rows for user=%s:%w", userID, error)
	}
	return &category, nil
}

func (r *CategoryRepository) GetCategories(ctx context.Context, userID string,
	query *category.GetCategoriesQuery,
) (*model.PaginatedResponse[category.Category], error) {
	stmt := `
		SELECT
			*
		FROM
			todo_categories
		WHERE
			user_id=@user_id
	`

	args := pgx.NamedArgs{
		"user_id": userID,
	}

	// Add search filter if provided
	if query.Search != nil {
		stmt += ` AND name ILIKE '%' || @search || '%'`
		args["search"] = *query.Search
	}

	// Add sorting
	sortColumn := "name"
	if query.Sort != nil {
		sortColumn = *query.Sort
	}
	sortOrder := "asc"
	if query.Order != nil {
		sortOrder = *query.Order
	}
	stmt += fmt.Sprintf(" ORDER BY %s %s", sortColumn, sortOrder)

	// Add pagination
	stmt += ` LIMIT @limit OFFSET @offset`
	args["limit"] = *query.Limit
	args["offset"] = (*query.Page - 1) * (*query.Limit)

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get categories query for user_id=%s: %w", userID, err)
	}

	categories, err := pgx.CollectRows(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PaginatedResponse[category.Category]{
				Data:       []category.Category{},
				Page:       *query.Page,
				Limit:      *query.Limit,
				Total:      0,
				TotalPages: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to collect rows from table:todo_categories for user_id=%s: %w", userID, err)
	}

	// Get total count
	countStmt := `
		SELECT
			COUNT(*)
		FROM
			todo_categories
		WHERE
			user_id=@user_id
	`

	countArgs := pgx.NamedArgs{
		"user_id": userID,
	}

	if query.Search != nil {
		countStmt += ` AND name ILIKE '%' || @search || '%'`
		countArgs["search"] = *query.Search
	}

	var total int
	err = r.server.DB.Pool.QueryRow(ctx, countStmt, countArgs).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count of categories for user_id=%s: %w", userID, err)
	}

	return &model.PaginatedResponse[category.Category]{
		Data:       categories,
		Page:       *query.Page,
		Limit:      *query.Limit,
		Total:      total,
		TotalPages: (total + *query.Limit - 1) / *query.Limit,
	}, nil
}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, userID string,
	categoryID uuid.UUID, payload *category.UpdateCategoryPayload,
) (*category.Category, error) {
	stmt := `UPDATE todo_categories SET `
	args := pgx.NamedArgs{
		"id":      categoryID,
		"user_id": userID,
	}
	setClauses := []string{}

	if payload.Name != nil {
		setClauses = append(setClauses, "name = @name")
		args["name"] = *payload.Name
	}
	if payload.Color != nil {
		setClauses = append(setClauses, "color = @color")
		args["color"] = *payload.Color
	}
	if payload.Description != nil {
		setClauses = append(setClauses, "description = @description")
		args["description"] = *payload.Description
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	stmt += strings.Join(setClauses, ", ")
	stmt += ` WHERE id = @id AND user_id = @user_id RETURNING *`

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute update category query for category_id=%s user_id=%s: %w", categoryID.String(), userID, err)
	}

	categoryItem, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[category.Category])
	if err != nil {
		return nil, fmt.Errorf("failed to collect row from table:todo_categories for category_id=%s user_id=%s: %w", categoryID.String(), userID, err)
	}

	return &categoryItem, nil
}

func (r *CategoryRepository) DeleteCategory(ctx context.Context, userID string, categoryID uuid.UUID) error {
	result, err := r.server.DB.Pool.Exec(ctx, `
		DELETE FROM todo_categories
		WHERE id = @id AND user_id = @user_id
	`, pgx.NamedArgs{
		"id":      categoryID,
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
