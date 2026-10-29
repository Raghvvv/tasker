package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Raghvvv/tasker/internal/errs"
	"github.com/Raghvvv/tasker/internal/model"
	"github.com/Raghvvv/tasker/internal/model/todo"
	"github.com/Raghvvv/tasker/internal/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type TodoRepository struct {
	server *server.Server
}

func NewTodoRepository(s *server.Server) *TodoRepository {
	return &TodoRepository{server: s}
}

func (r *TodoRepository) CreateTodo(ctx context.Context, userID string, payload *todo.CreateTodoPayload) (*todo.Todo, error) {
	stmt := `
		INTSERT INTO
		todos(
			user_id,
			title,
			description,
			priority,
			due_date,
			metadata,
			category_id,
			parent_todo_id,

		)
		VALUES
		(
			@user_id,
			@title,
			@description,
			@priority,
			@due_date,
			@metadata,
			@category_id,
			@parent_todo_id,
		
		)
		RETURNING *
	`
	priority := todo.PriorityMedium
	if payload.Priority != nil {
		priority = *payload.Priority
	}
	rows, error := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id":        userID,
		"title":          payload.Title,
		"description":    payload.Description,
		"priority":       priority,
		"due_date":       payload.DueDate,
		"parent_todo_id": payload.ParentTodoID,
		"category_id":    payload.CategoryID,
		"metadata":       payload.Metadata,
	})
	if error != nil {
		return nil, fmt.Errorf("error in running query for userId=%s, title=%s: %w", userID, payload.Title, error)
	}
	todoItem, error := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.Todo])
	if error != nil {
		return nil, fmt.Errorf("failed to collect row for userId=%s, title=%s:%w", userID, payload.Title, error)
	}
	return &todoItem, nil

}

func (r *TodoRepository) GetTodoByID(ctx context.Context, userID string, todoID uuid.UUID) (*todo.Populatedtodo, error) {
	stmt := `
	SELECT
		t.*,
		CASE
			WHEN c.id IS NOT NULL THEN to_jsonb(camel (c))
			ELSE NULL
		END AS category,
		COALESCE(
			jsonb_agg(
				to_jsonb(camel (child))
				ORDER BY
					child.sort_order ASC,
					child.created_at ASC
			) FILTER (
				WHERE
					child.id IS NOT NULL
			),
			'[]'::JSONB
		) AS children,
		COALESCE(
			jsonb_agg(
				to_jsonb(camel (com))
				ORDER BY
					com.created_at ASC
			) FILTER (
				WHERE
					com.id IS NOT NULL
			),
			'[]'::JSONB
		) AS comments,
		
	FROM
		todos t
		LEFT JOIN todo_categories c ON c.id=t.category_id
		AND c.user_id=@user_id
		LEFT JOIN todos child ON child.parent_todo_id=t.id
		AND child.user_id=@user_id
		LEFT JOIN todo_comments com ON com.todo_id=t.id
		AND com.user_id=@user_id
		LEFT JOIN todo_attachments att ON att.todo_id=t.id
	WHERE
		t.id=@id
		AND t.user_id=@user_id
	GROUP BY
		t.id,
		c.id `

	// query returns a list of rows
	rows, error := r.server.DB.Pool.Query(ctx, stmt, pgx.NamedArgs{
		"user_id": userID,
		"id":      todoID,
	})

	if error != nil {
		return nil, fmt.Errorf("failed to extract rows for user=%s, todoID=%s:%w", userID, todoID.String(), error)
	}
	//collectRows formats the rows to a struct
	todoItems, error := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.Populatedtodo])
	if error != nil {
		return nil, fmt.Errorf("failed to collect todoItem from rows for userID=%s,todoID=%s:%w", userID, todoID.String(), error)
	}
	return &todoItems, nil

}

func (r *TodoRepository) GetTodos(ctx context.Context, userID string, query *todo.GetTodoQuery) (*model.PaginatedResponse[todo.Populatedtodo], error) {
	stmt := `
		SELECT
			t.*,
			CASE
				WHEN c.id IS NOT NULL THEN to_jsonb(camel (c))
				ELSE NULL
			END AS category,
			COALESCE(
				jsonb_agg(
					to_jsonb(camel (child))
					ORDER BY
						child.sort_order ASC,
						child.created_at ASC
				) FILTER (
					WHERE
						child.id IS NOT NULL
				),
				'[]'::JSONB
			) AS children,
			COALESCE(
				jsonb_agg(
					to_jsonb(camel (com))
					ORDER BY
						com.created_at ASC
				) FILTER (
					WHERE
						com.id IS NOT NULL
				),
				'[]'::JSONB
			) AS comments,
			
		FROM
			todos t
			LEFT JOIN todo_categories c ON c.id=t.category_id
			AND c.user_id=@user_id
			LEFT JOIN todos child ON child.parent_todo_id=t.id
			AND child.user_id=@user_id
			LEFT JOIN todo_comments com ON com.todo_id=t.id
			AND com.user_id=@user_id
			
		WHERE
			t.id=@id
			AND t.user_id=@user_id
		GROUP BY
			t.id,
			c.id `

	args := pgx.NamedArgs{
		"user_id": userID,
	}
	conditions := []string{"t.user_id=@userID"}

	if query.Status != nil {
		conditions = append(conditions, "t.status=@status")
		args["status"] = *query.Status

	}
	if query.Priority != nil {
		conditions = append(conditions, "t.priority=@priority")
		args["priority"] = *query.Priority
	}
	if query.CategoryID != nil {
		conditions = append(conditions, "t.category_id=@category_id")
		args["category_id"] = *query.CategoryID
	}
	if query.ParentTodoID != nil {
		conditions = append(conditions, "t.parent_todo_id=@parent_todo_id")
		args["parent_todo_id"] = *query.ParentTodoID
	} else {
		// By default, only show root todos (no parent)
		conditions = append(conditions, "t.parent_todo_id IS NULL")
	}
	if query.Completed != nil {
		conditions = append(conditions, "t.completed=@completed")
		args["completed"] = query.Completed
	}
	if query.DueFrom != nil {
		conditions = append(conditions, "due_from>=@due_from")
		args["due_from"] = *query.DueFrom
	}
	if query.DueTo != nil {
		conditions = append(conditions, "due_to<=@due_to")
		args["due_to"] = *query.DueTo
	}
	if query.Overdue != nil {
		conditions = append(conditions, "t.due_date>=NOW() AND t.status!='completed'")

	}
	if query.Completed != nil {
		if *query.Completed {
			conditions = append(conditions, "t.status='completed'")
		} else {
			conditions = append(conditions, "t.status!='completed")
		}
	}
	if query.Search != nil {
		conditions = append(conditions, "t.title ILIKE @search OR t.description ILIKE @search")
		args["search"] = "%" + *query.Search + "%"
	}
	if len(conditions) > 0 {
		stmt += "WHERE" + strings.Join(conditions, "AND")
	}
	countstmt := "SELECT COUNT (*) FROM todos t "
	var count int
	err := r.server.DB.Pool.QueryRow(ctx, countstmt, args).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve count of rows for userID=%s,%w", userID, err)
	}
	stmt += "GROUP BY t.id,c.id"
	if query.Sort != nil {
		stmt += " ORDER BY t." + *query.Sort
		if query.Order != nil && *query.Order == "desc" {
			stmt += " DESC"
		} else {
			stmt += " ASC"
		}
	} else {
		stmt += " ORDER BY t.created_at DESC"
	}
	stmt += " LIMIT @limit OFFSET @offset"
	args["limit"] = *query.Limit
	args["offset"] = (*query.Page - 1) * (*query.Limit)

	rows, err := r.server.DB.Pool.Query(ctx, stmt, args)
	if err != nil {
		return nil, fmt.Errorf("failed to execute get todos query for user_id=%s: %w", userID, err)
	}

	todos, err := pgx.CollectRows(rows, pgx.RowToStructByName[todo.Populatedtodo])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &model.PaginatedResponse[todo.Populatedtodo]{
				Data:       []todo.Populatedtodo{},
				Page:       *query.Page,
				Limit:      *query.Limit,
				Total:      0,
				TotalPages: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to collect rows from table:todos for user_id=%s: %w", userID, err)
	}

	return &model.PaginatedResponse[todo.Populatedtodo]{
		Data:       todos,
		Page:       *query.Page,
		Limit:      *query.Limit,
		Total:      count,
		TotalPages: (count + *query.Limit - 1) / *query.Limit,
	}, nil
}

func (r *TodoRepository) UpdateTodo(ctx context.Context, userID string, payload *todo.UpdateTodoPayload) (*todo.Todo, error) {
	stmt := `UPDATE todos t SET`
	setClauses := []string{}
	args := pgx.NamedArgs{
		"user_id": userID,
		"todo_id": payload.ID,
	}

	if payload.Title != nil {
		setClauses = append(setClauses, "title=@title")
		args["title"] = *payload.Title
	}
	if payload.Description != nil {
		setClauses = append(setClauses, "description=@description")
		args["description"] = *payload.Description
	}
	if payload.Status != nil {
		setClauses = append(setClauses, "status=@status")
		args["status"] = *payload.Status
	}
	if payload.Priority != nil {
		setClauses = append(setClauses, "priority=@priority")
		args["priority"] = *payload.Priority
	}
	if payload.DueDate != nil {
		setClauses = append(setClauses, "due_date=@dueDate")
		args["dueDate"] = *payload.DueDate
	}
	if payload.ParentTodoID != nil {
		setClauses = append(setClauses, "parent_todo_id=@parentTodoID")
		args["parentTodoID"] = *payload.ParentTodoID
	}
	if payload.CategoryID != nil {
		setClauses = append(setClauses, "category_id=@categoryID")
		args["categoryID"] = *payload.CategoryID
	}
	if payload.Metadata != nil {
		setClauses = append(setClauses, "metadata=@metadata")
		args["metadata"] = payload.Metadata
	}
	if len(setClauses) == 0 {
		return nil, errs.NewBadRequestError("no fields to update!!", false, nil, nil, nil)
	}
	stmt += strings.Join(setClauses, ",")
	stmt += "WHERE id=@todo_id AND user_id=@user_id RETURNING*"

	rows, error := r.server.DB.Pool.Query(ctx, stmt, args)
	if error != nil {
		return nil, fmt.Errorf("failed to execute query for todoID=%s, userID=%s:%w", payload.ID, userID, error)
	}
	updatedTodoItem, error := pgx.CollectOneRow(rows, pgx.RowToStructByName[todo.Todo])
	if error != nil {
		return nil, fmt.Errorf("failed to populate struct from rows %w", error)
	}
	return &updatedTodoItem, nil

}

func (r *TodoRepository) DeleteTodo(ctx context.Context, userID string, todoID uuid.UUID) error {
	stmt := `
		DELETE FROM todos
		WHERE
			id=@todo_id
			AND user_id=@user_id
	`

	result, err := r.server.DB.Pool.Exec(ctx, stmt, pgx.NamedArgs{
		"todo_id": todoID,
		"user_id": userID,
	})
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if result.RowsAffected() == 0 {
		code := "TODO_NOT_FOUND"
		return errs.NewNotFoundError("todo not found", false, &code)
	}

	return nil
}
