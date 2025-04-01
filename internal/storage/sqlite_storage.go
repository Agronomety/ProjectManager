package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Agronomety/ProjectManager/internal/models"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {

	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	err = createTables(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	return &SQLiteStorage{db: db}, nil
}

func createTables(db *sql.DB) error {

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			path TEXT NOT NULL UNIQUE,
			description TEXT,
			readme_path TEXT,
			last_opened DATETIME,
			tags TEXT,
			icon TEXT
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create projects table: %v", err)
	}

	return nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

func (s *SQLiteStorage) Create(project *models.Project) error {
	query := `
		INSERT INTO projects 
		(name, path, description, readme_path, last_opened, tags, icon) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	tagsStr := ""
	if len(project.Tags) > 0 {
		tagsStr = strings.Join(project.Tags, ",")
	}

	result, err := s.db.Exec(
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

func (s *SQLiteStorage) Update(project *models.Project) error {
	query := `
		UPDATE projects 
		SET name = ?, description = ?, readme_path = ?, last_opened = ?, tags = ?, icon = ?
		WHERE id = ?
	`

	tagsStr := ""
	if len(project.Tags) > 0 {
		tagsStr = strings.Join(project.Tags, ",")
	}

	_, err := s.db.Exec(
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

func (s *SQLiteStorage) Delete(id int64) error {
	_, err := s.db.Exec("DELETE FROM projects WHERE id = ?", id)
	return err
}

func (s *SQLiteStorage) GetByID(id int64) (*models.Project, error) {
	query := `
		SELECT id, name, path, description, readme_path, last_opened, tags, icon 
		FROM projects 
		WHERE id = ?
	`

	var project models.Project
	var tagsStr string

	err := s.db.QueryRow(query, id).Scan(
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

	if tagsStr != "" {
		project.Tags = strings.Split(tagsStr, ",")
	}

	return &project, nil
}

func (s *SQLiteStorage) ListAll() ([]models.Project, error) {
	query := `
		SELECT id, name, path, description, readme_path, last_opened, tags, icon 
		FROM projects
	`

	rows, err := s.db.Query(query)
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

	return projects, nil
}
