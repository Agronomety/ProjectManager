package service

import (
	"strings"

	"github.com/Agronomety/ProjectManager/internal/models"
	"github.com/Agronomety/ProjectManager/internal/storage"
)

type ProjectService interface {
	CreateProject(project *models.Project) error
	UpdateProject(project *models.Project) error
	DeleteProject(id int64) error
	GetProject(id int64) (*models.Project, error)
	ListProjects() ([]models.Project, error)
	SearchProjects(query string) ([]models.Project, error)
}

type DefaultProjectService struct {
	repo storage.ProjectRepository
}

func NewProjectService(repo storage.ProjectRepository) ProjectService {
	return &DefaultProjectService{repo: repo}
}

func (s *DefaultProjectService) CreateProject(project *models.Project) error {
	return s.repo.Create(project)
}

func (s *DefaultProjectService) UpdateProject(project *models.Project) error {
	return s.repo.Update(project)
}

func (s *DefaultProjectService) DeleteProject(id int64) error {
	return s.repo.Delete(id)
}

func (s *DefaultProjectService) GetProject(id int64) (*models.Project, error) {
	return s.repo.GetByID(id)
}

func (s *DefaultProjectService) ListProjects() ([]models.Project, error) {
	return s.repo.ListAll()
}

func (s *DefaultProjectService) SearchProjects(query string) ([]models.Project, error) {
	projects, err := s.repo.ListAll()
	if err != nil {
		return nil, err
	}

	// Simple case-insensitive search
	var results []models.Project
	query = strings.ToLower(query)
	for _, project := range projects {
		if strings.Contains(strings.ToLower(project.Name), query) ||
			strings.Contains(strings.ToLower(project.Description), query) {
			results = append(results, project)
		}
	}

	return results, nil
}
