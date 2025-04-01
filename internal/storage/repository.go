package storage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Agronomety/ProjectManager/internal/models"
)

type ProjectRepository interface {
	Create(project *models.Project) error
	Update(project *models.Project) error
	Delete(id int64) error
	GetByID(id int64) (*models.Project, error)
	ListAll() ([]models.Project, error)
}

func (r *SQLiteProjectRepository) Create(project *models.Project) error {

	query := `
 		INSERT INTO projects
		(name, path, description, readme_path, last_opened, tags, icon)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	tagsStr := strings.Join(project.Tags, ",")

	result, err := r.db.Exec(
		query,
		project.Name,
		project.Path,
		project.Description,
		project.ReadmePath,
		project.LastOpened,
		tagsStr,
		project.Icon,
	)

	if err != nil {
		return fmt.Errorf("failed to insert project: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %v", err)
	}

	project.ID = id

	return nil
}

func (r *SQLiteProjectRepository) Update(project *models.Project) error {
	query := `
		UPDATE projects
		SET name = ?, description = ?, readme_path = ?, last_opened = ?, tags = ?, icon = ?
		WHERE id = ?
	`

	tagsStr := strings.Join(project.Tags, ",")

	_, err := r.db.Exec(
		query,
		project.Name,
		project.Description,
		project.ReadmePath,
		project.LastOpened,
		tagsStr,
		project.Icon,
		project.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update project: %v", err)
	}

	return nil
}

func (r *SQLiteProjectRepository) Delete(id int64) error {
	query := `
		DELETE FROM projects
		WHERE id = ?
	`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %v", err)
	}

	return nil
}

func (r *SQLiteProjectRepository) GetByID(id int64) (*models.Project, error) {
	query := `
		SELECT id, name, path, description, readme_path, last_opened, tags, icon
		FROM projects
		WHERE id = ?
	`

	var project models.Project
	var tagsStr string

	err := r.db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Path,
		&project.Description,
		&project.ReadmePath,
		&project.LastOpened,
		&tagsStr,
		&project.Icon,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get project: %v", err)
	}

	project.Tags = strings.Split(tagsStr, ",")

	return &project, nil
}

func (r *SQLiteProjectRepository) ListAll() ([]models.Project, error) {
	query := `
        SELECT id, name, path, description, readme_path, 
               last_opened, tags, icon 
        FROM projects
    `

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %v", err)
	}
	defer rows.Close()

	var projects []models.Project

	for rows.Next() {
		var project models.Project
		var tagsStr string

		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Path,
			&project.Description,
			&project.ReadmePath,
			&project.LastOpened,
			&tagsStr,
			&project.Icon,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %v", err)
		}

		if tagsStr != "" {
			project.Tags = strings.Split(tagsStr, ",")
		}

		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error reading projects: %v", err)
	}

	return projects, nil
}

type SQLiteProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(storage *SQLiteStorage) ProjectRepository {
	return &SQLiteProjectRepository{db: storage.db}
}
