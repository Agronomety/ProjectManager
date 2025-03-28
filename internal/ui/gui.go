package ui

import (
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Agronomety/ProjectManager/internal/models"
	"github.com/Agronomety/ProjectManager/internal/service"
	"github.com/Agronomety/ProjectManager/pkg/utils"
	"github.com/Agronomety/ProjectManager/pkg/vscode"
)

type ProjectManagerUI struct {
	app                  fyne.App
	window               fyne.Window
	projectService       service.ProjectService
	projectList          *widget.List
	projectDetails       *widget.Form
	descriptionEdit      *widget.Entry
	readmeViewer         *widget.Label
	currentProjects      []models.Project
	vsCodeLauncher       *vscode.Launcher
	selectedProjectIndex int
}

func NewProjectManagerUI(projectService service.ProjectService) *ProjectManagerUI {
	a := app.New()
	w := a.NewWindow("Project Manager")
	w.Resize(fyne.NewSize(1200, 800))

	ui := &ProjectManagerUI{
		app:            a,
		window:         w,
		projectService: projectService,
		vsCodeLauncher: vscode.NewLauncher(projectService),
	}

	ui.createUI()
	return ui
}

func (ui *ProjectManagerUI) createUI() {
	// Create a banner with Go-themed styling
	bannerLabel := widget.NewLabel("ProjectManager")
	bannerLabel.TextStyle = fyne.TextStyle{
		Bold: true,
	}

	// Create a blue rectangle for the banner background
	// Using Go's light blue color (#00ADD8)
	bannerBg := canvas.NewRectangle(color.NRGBA{R: 0, G: 173, B: 216, A: 255})
	bannerBg.SetMinSize(fyne.NewSize(200, 40))

	// Overlay the label on the banner background
	bannerContainer := container.NewStack(
		bannerBg,
		container.NewCenter(bannerLabel),
	)

	// Initialize project list
	ui.projectList = widget.NewList(
		func() int { return len(ui.currentProjects) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Project Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			label := item.(*widget.Label)
			if id < len(ui.currentProjects) {
				label.SetText(ui.currentProjects[id].Name)
			}
		},
	)

	// Create Project Buttons
	newProjectBtn := widget.NewButton("New Project", ui.showNewProjectDialog)
	importProjectBtn := widget.NewButton("Import Projects", ui.showImportProjectsDialog)

	// Wrap buttons in a container
	buttonContainer := container.NewVBox(
		newProjectBtn,
		importProjectBtn,
	)

	// Project List Container with banner on top
	projectListContainer := container.NewBorder(
		container.NewVBox(bannerContainer, buttonContainer), // Top - banner and buttons
		nil,            // Bottom
		nil,            // Left
		nil,            // Right
		ui.projectList, // Center
	)

	// Initialize description edit
	ui.descriptionEdit = widget.NewMultiLineEntry()
	ui.descriptionEdit.SetPlaceHolder("Enter project description...")

	// Initialize README viewer
	ui.readmeViewer = widget.NewLabel("No README loaded")
	ui.readmeViewer.Wrapping = fyne.TextWrapWord

	// Create a scrollable container for README content
	readmeScrollContainer := container.NewScroll(ui.readmeViewer)
	// Set a minimum size for the scroll container so it displays properly
	readmeScrollContainer.SetMinSize(fyne.NewSize(400, 300))

	// README upload button
	readmeUploadBtn := widget.NewButton("Upload README", ui.uploadReadmeFile)

	//Open in VSCode button
	openInVSCodeBtn := widget.NewButton("Open in VSCode", func() {
		if ui.selectedProjectIndex < 0 || ui.selectedProjectIndex >= len(ui.currentProjects) {
			dialog.ShowError(fmt.Errorf("no project selected"), ui.window)
			return
		}
		project := ui.currentProjects[ui.selectedProjectIndex]
		err := ui.vsCodeLauncher.OpenProject(&project)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to open project in VSCode: %v", err), ui.window)
			return
		}
		project.LastOpened = time.Now()
		err = ui.projectService.UpdateProject(&project)
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to update last opened time: %v", err), ui.window)
			return
		}
	})

	// Remove Project button with confirmation dialog
	removeProjectBtn := widget.NewButton("Remove Project", func() {
		if ui.selectedProjectIndex < 0 || ui.selectedProjectIndex >= len(ui.currentProjects) {
			dialog.ShowError(fmt.Errorf("no project selected"), ui.window)
			return
		}

		project := ui.currentProjects[ui.selectedProjectIndex]

		// Add confirmation dialog
		confirmDialog := dialog.NewConfirm(
			"Confirm Removal",
			fmt.Sprintf("Are you sure you want to remove project '%s'? This action cannot be undone.", project.Name),
			func(confirmed bool) {
				if !confirmed {
					return // User canceled the operation
				}

				// User confirmed, proceed with deletion
				err := ui.projectService.DeleteProject(project.ID)
				if err != nil {
					dialog.ShowError(fmt.Errorf("failed to remove project: %v", err), ui.window)
					return
				}

				ui.currentProjects = append(ui.currentProjects[:ui.selectedProjectIndex], ui.currentProjects[ui.selectedProjectIndex+1:]...)
				ui.projectList.Refresh()
				ui.selectedProjectIndex = -1              // Reset selection
				ui.updateProjectDetails(models.Project{}) // Clear details
			},
			ui.window,
		)

		// Show the confirmation dialog
		confirmDialog.Show()
	})

	// Project Details Form - Update the README item to use the scroll container
	ui.projectDetails = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Project Name", Widget: widget.NewLabel("")},
			{Text: "Description", Widget: ui.descriptionEdit},
			{Text: "README", Widget: readmeUploadBtn},
			{Text: "Open in VSCode", Widget: openInVSCodeBtn},
			{Text: "Remove Project", Widget: removeProjectBtn},
			{Text: "README Viewer", Widget: readmeScrollContainer}, // Use the scroll container here
		},
	}

	// Make the form scrollable too, for projects with lots of metadata or description
	formScroll := container.NewScroll(ui.projectDetails)

	// Project List Selection Handler
	ui.projectList.OnSelected = func(id widget.ListItemID) {
		if id < len(ui.currentProjects) {
			ui.selectedProjectIndex = id
			project := ui.currentProjects[id]
			ui.updateProjectDetails(project)
		}
	}

	// Main Layout
	split := container.NewHSplit(
		projectListContainer,
		formScroll, // Use the scrollable form instead of direct form
	)
	split.Offset = 0.3 // 30% list, 70% details

	// Set window content
	ui.window.SetContent(split)

	// Load projects
	ui.loadProjects()
}

func (ui *ProjectManagerUI) showNewProjectDialog() {
	// Project path selection
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("Select project directory")

	// Project name entry
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter project name")

	// Description entry
	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Enter project description (optional)")

	// Path selection button
	pathSelectBtn := widget.NewButton("Browse", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, ui.window)
				return
			}
			if uri == nil {
				return
			}
			pathEntry.SetText(uri.Path())

			// Auto-fill project name if not already set
			if nameEntry.Text == "" {
				nameEntry.SetText(utils.GetProjectName(uri.Path()))
			}
		}, ui.window)
	})

	// Create content for the dialog
	content := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Project Path", Widget: container.NewHBox(pathEntry, pathSelectBtn)},
			{Text: "Project Name", Widget: nameEntry},
			{Text: "Description", Widget: descriptionEntry},
		},
	}

	// Create dialog
	dialog.ShowCustomConfirm("Create New Project", "Create", "Cancel", content, func(b bool) {
		if !b {
			return // Cancel pressed
		}

		// Validate inputs
		projectPath := pathEntry.Text
		projectName := nameEntry.Text

		if projectPath == "" || projectName == "" {
			dialog.ShowError(fmt.Errorf("project path and name are required"), ui.window)
			return
		}

		// Validate project path
		err := utils.ValidateProjectPath(projectPath)
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		// Try to read README
		readmeContent, _ := utils.ReadReadmeFile(projectPath)

		// Create project object
		project := &models.Project{
			Name:        projectName,
			Path:        projectPath,
			Description: descriptionEntry.Text,
			LastOpened:  time.Now(),
			ReadmePath:  "", // We'll add this if a README is found
		}

		// If README found, save it
		if readmeContent != "" {
			readmePath := filepath.Join(projectPath, "README.md")
			err = ioutil.WriteFile(readmePath, []byte(readmeContent), 0644)
			if err == nil {
				project.ReadmePath = readmePath
			}
		}

		// Attempt to get project metadata
		metadata := utils.ScanProjectMetadata(projectPath)
		if len(metadata) > 0 {
			// You could potentially extract tags or other info from metadata
			project.Tags = []string{} // Add logic to extract tags if needed
		}

		// Save project
		err = ui.projectService.CreateProject(project)
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		// Refresh project list
		ui.loadProjects()
	}, ui.window)
}

// / Show import projects dialog
// This function allows the user to select multiple directories and import projects from them
func (ui *ProjectManagerUI) showImportProjectsDialog() {
	// Allow selecting multiple directories
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		if uri == nil {
			return
		}

		basePath := uri.Path()

		// Find potential project roots
		projectPaths, err := utils.FindProjectRoots([]string{basePath})
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		// Confirm project import
		confirmImport := dialog.NewConfirm(
			"Import Projects",
			fmt.Sprintf("Found %d potential projects. Import all?", len(projectPaths)),
			func(confirmed bool) {
				if !confirmed {
					return
				}

				// Import projects
				var importedProjects []*models.Project
				var errors []error

				for _, path := range projectPaths {
					project := &models.Project{
						Name:       utils.GetProjectName(path),
						Path:       path,
						LastOpened: time.Now(),
					}

					// Try to read README
					readmeContent, _ := utils.ReadReadmeFile(path)
					if readmeContent != "" {
						readmePath := filepath.Join(path, "README.md")
						err = ioutil.WriteFile(readmePath, []byte(readmeContent), 0644)
						if err == nil {
							project.ReadmePath = readmePath
						}
					}

					// Scan metadata
					metadata := utils.ScanProjectMetadata(path)
					if len(metadata) > 0 {
						// You could potentially extract additional info
						project.Tags = extractTagsFromMetadata(metadata)
					}

					// Create project
					err := ui.projectService.CreateProject(project)
					if err != nil {
						errors = append(errors, fmt.Errorf("failed to import %s: %v", path, err))
					} else {
						importedProjects = append(importedProjects, project)
					}
				}

				// Show import results
				if len(errors) > 0 {
					errorMsg := "Some projects failed to import:\n"
					for _, e := range errors {
						errorMsg += e.Error() + "\n"
					}
					dialog.ShowError(fmt.Errorf(errorMsg), ui.window)
				}

				// Refresh project list
				ui.loadProjects()
			},
			ui.window,
		)
		confirmImport.Show()
	}, ui.window)
}

// Helper function to extract tags from project metadata
func extractTagsFromMetadata(metadata map[string]string) []string {
	var tags []string

	// Example: Extract language tags
	if _, exists := metadata["go.mod"]; exists {
		tags = append(tags, "Go")
	}
	if _, exists := metadata["package.json"]; exists {
		tags = append(tags, "JavaScript", "Node.js")
	}
	if _, exists := metadata["pyproject.toml"]; exists {
		tags = append(tags, "Python")
	}
	if _, exists := metadata["pom.xml"]; exists {
		tags = append(tags, "Java", "Maven")
	}

	return tags
}

func (ui *ProjectManagerUI) updateProjectDetails(project models.Project) {
	// Update project name
	ui.projectDetails.Items[0].Widget.(*widget.Label).SetText(project.Name)

	// Update description
	ui.descriptionEdit.SetText(project.Description)

	// Load README
	if project.ReadmePath != "" {
		content, err := ioutil.ReadFile(project.ReadmePath)
		if err != nil {
			ui.readmeViewer.SetText("Error reading README")
		} else {
			ui.readmeViewer.SetText(string(content))
		}
	} else {
		ui.readmeViewer.SetText("No README loaded")
	}
}

func (ui *ProjectManagerUI) uploadReadmeFile() {
	dialog.ShowFileOpen(func(uc fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		if uc == nil {
			return
		}
		defer uc.Close()

		// Get the selected file's path
		filePath := uc.URI().Path()

		// Get the currently selected project
		selectedIndex := ui.selectedProjectIndex
		if selectedIndex < 0 || selectedIndex >= len(ui.currentProjects) {
			dialog.ShowError(fmt.Errorf("no project selected"), ui.window)
			return
		}

		project := ui.currentProjects[selectedIndex]
		project.ReadmePath = filePath

		// Update project in service
		err = ui.projectService.UpdateProject(&project)
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		// Refresh project details
		ui.updateProjectDetails(project)
	}, ui.window)
}

func (ui *ProjectManagerUI) loadProjects() {
	projects, err := ui.projectService.ListProjects()
	if err != nil {
		log.Printf("Error loading projects: %v", err)
		return
	}

	ui.currentProjects = projects

	// Refresh the project list
	if ui.projectList != nil {
		ui.projectList.Refresh()
	}

	// If there are projects, select the first one
	if len(projects) > 0 {
		ui.projectList.Select(0)
	}
}

func (ui *ProjectManagerUI) Run() {
	ui.window.ShowAndRun()
}
