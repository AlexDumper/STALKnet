package repository

import "database/sql"

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// GetTasks возвращает все задачи с фильтрацией
func (r *TaskRepository) GetTasks(status string, roomID int) ([]Task, error) {
	query := `
		SELECT id, title, description, creator_id, assignee_id, room_id, status, created_at, completed_at, confirmed_at
		FROM tasks
		WHERE ($1 = '' OR status = $1)
		AND ($2 = 0 OR room_id = $2)
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, status, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.CreatorID,
			&task.AssigneeID, &task.RoomID, &task.Status, &task.CreatedAt,
			&task.CompletedAt, &task.ConfirmedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// CreateTask создаёт новую задачу
func (r *TaskRepository) CreateTask(title, description string, creatorID, assigneeID, roomID int) (int, error) {
	var id int
	query := `INSERT INTO tasks (title, description, creator_id, assignee_id, room_id) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.db.QueryRow(query, title, description, creatorID, assigneeID, roomID).Scan(&id)
	return id, err
}

// GetTaskByID находит задачу по ID
func (r *TaskRepository) GetTaskByID(id int) (*Task, error) {
	query := `
		SELECT id, title, description, creator_id, assignee_id, room_id, status, created_at, completed_at, confirmed_at
		FROM tasks WHERE id = $1
	`
	row := r.db.QueryRow(query, id)

	var task Task
	err := row.Scan(
		&task.ID, &task.Title, &task.Description, &task.CreatorID,
		&task.AssigneeID, &task.RoomID, &task.Status, &task.CreatedAt,
		&task.CompletedAt, &task.ConfirmedAt,
	)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

// UpdateTask обновляет задачу
func (r *TaskRepository) UpdateTask(id int, title, description string, assigneeID int, status string) error {
	query := `UPDATE tasks SET title = $1, description = $2, assignee_id = $3, status = $4 WHERE id = $5`
	_, err := r.db.Exec(query, title, description, assigneeID, status, id)
	return err
}

// CompleteTask отмечает задачу выполненной
func (r *TaskRepository) CompleteTask(id int) error {
	query := `UPDATE tasks SET status = 'done', completed_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// ConfirmTask подтверждает выполнение задачи
func (r *TaskRepository) ConfirmTask(id int) error {
	query := `UPDATE tasks SET status = 'confirmed', confirmed_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// DeleteTask удаляет задачу
func (r *TaskRepository) DeleteTask(id int) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// Task модель
type Task struct {
	ID          int
	Title       string
	Description string
	CreatorID   int
	AssigneeID  int
	RoomID      int
	Status      string
}
