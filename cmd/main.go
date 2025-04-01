package main

import (
	"log"

	"github.com/Agronomety/ProjectManager/internal/config"
	"github.com/Agronomety/ProjectManager/internal/service"
	"github.com/Agronomety/ProjectManager/internal/storage"
	"github.com/Agronomety/ProjectManager/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := storage.NewSQLiteStorage(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	projectRepo := storage.NewProjectRepository(db)
	projectService := service.NewProjectService(projectRepo)

	app := ui.NewProjectManagerUI(projectService)
	app.Run()
}
