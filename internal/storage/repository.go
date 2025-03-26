package storage

import (
	"database/sql"

	"github.com/Agronomety/ProjectManager/internal/models"
)

type ProjectRepository interface {
	Create(project *models.Project) error
	Update(project *models.Project) error
	Delete(id int64) error
	GetByID(id int64) (*models.Project, error)
	ListAll() ([]models.Project, error)
}

type SQLiteProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &SQLiteProjectRepository{db: db}
}
