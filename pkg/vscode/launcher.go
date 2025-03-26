package vscode

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Agronomety/ProjectManager/internal/models"
	"github.com/Agronomety/ProjectManager/internal/service"
)

type Launcher struct {
	projectService service.ProjectService
}

func NewLauncher(projectService service.ProjectService) *Launcher {
	return &Launcher{
		projectService: projectService,
	}
}

func (l *Launcher) OpenProject(project *models.Project) error {
	// Verify project path exists
	if _, err := os.Stat(project.Path); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %s", project.Path)
	}

	// Prepare VS Code launch command based on operating system
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "code", project.Path)
	case "darwin": // macOS
		cmd = exec.Command("open", "-a", "Visual Studio Code", project.Path)
	default: // Linux and other Unix-like systems
		cmd = exec.Command("code", project.Path)
	}

	// Update last opened time
	project.LastOpened = time.Now()
	err := l.projectService.UpdateProject(project)
	if err != nil {
		log.Printf("Failed to update last opened time: %v", err)
	}

	// Launch VS Code
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to launch VS Code: %v", err)
	}

	return nil
}

// Optional: Utility method to detect VS Code installation
func (l *Launcher) IsVSCodeInstalled() bool {
	_, err := exec.LookPath("code")
	return err == nil
}

// Optional: Find VS Code installation path
func (l *Launcher) FindVSCodePath() (string, error) {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		paths = []string{
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "Microsoft VS Code", "Code.exe"),
			filepath.Join(os.Getenv("ProgramFiles"), "Microsoft VS Code", "Code.exe"),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Microsoft VS Code", "Code.exe"),
		}
	case "darwin":
		paths = []string{
			"/Applications/Visual Studio Code.app/Contents/MacOS/Electron",
			"/Applications/VSCode.app/Contents/MacOS/Electron",
		}
	default: // Linux
		paths = []string{
			"/usr/bin/code",
			"/usr/local/bin/code",
			filepath.Join(os.Getenv("HOME"), ".local", "bin", "code"),
		}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("VS Code installation not found")
}
