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
	searchEntry          *widget.Entry
	currentProjects      []models.Project
	vsCodeLauncher       *vscode.Launcher
	selectedProjectIndex int
}

// NewProjectManagerUI creates and initializes a new project manager UI
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

// createUI builds the user interface components and layout
func (ui *ProjectManagerUI) createUI() {

	bannerLabel := widget.NewLabel("ProjectManager")
	bannerLabel.TextStyle = fyne.TextStyle{
		Bold: true,
	}

	bannerBg := canvas.NewRectangle(color.NRGBA{R: 0, G: 173, B: 216, A: 255})
	bannerBg.SetMinSize(fyne.NewSize(200, 40))

	bannerContainer := container.NewStack(
		bannerBg,
		container.NewCenter(bannerLabel),
	)

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

	newProjectBtn := widget.NewButton("New Project", ui.showNewProjectDialog)
	importProjectBtn := widget.NewButton("Import Projects", ui.showImportProjectsDialog)

	buttonContainer := container.NewVBox(
		newProjectBtn,
		importProjectBtn,
	)

	ui.searchEntry = widget.NewEntry()
	ui.searchEntry.SetPlaceHolder("Search projects...")
	searchIcon := widget.NewButton("🔍", func() {
		ui.performSearch(ui.searchEntry.Text)
	})
	searchBar := container.NewBorder(nil, nil, nil, searchIcon, ui.searchEntry)

	ui.searchEntry.OnSubmitted = func(query string) {
		ui.performSearch(query)
	}

	projectListContainer := container.NewBorder(
		container.NewVBox(bannerContainer, buttonContainer, searchBar), // Top - banner and buttons
		nil,            // Bottom
		nil,            // Left
		nil,            // Right
		ui.projectList, // Center
	)

	ui.descriptionEdit = widget.NewMultiLineEntry()
	ui.descriptionEdit.SetPlaceHolder("Enter project description...")

	ui.readmeViewer = widget.NewLabel("No README loaded")
	ui.readmeViewer.Wrapping = fyne.TextWrapWord

	readmeScrollContainer := container.NewScroll(ui.readmeViewer)
	readmeScrollContainer.SetMinSize(fyne.NewSize(400, 300))

	readmeUploadBtn := widget.NewButton("Upload README", ui.uploadReadmeFile)

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

	removeProjectBtn := widget.NewButton("Remove Project", func() {
		if ui.selectedProjectIndex < 0 || ui.selectedProjectIndex >= len(ui.currentProjects) {
			dialog.ShowError(fmt.Errorf("no project selected"), ui.window)
			return
		}

		project := ui.currentProjects[ui.selectedProjectIndex]

		confirmDialog := dialog.NewConfirm(
			"Confirm Removal",
			fmt.Sprintf("Are you sure you want to remove project '%s'? This action cannot be undone.", project.Name),
			func(confirmed bool) {
				if !confirmed {
					return
				}

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

		confirmDialog.Show()
	})

	ui.projectDetails = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Project Name", Widget: widget.NewLabel("")},
			{Text: "Description", Widget: ui.descriptionEdit},
			{Text: "README", Widget: readmeUploadBtn},
			{Text: "Open in VSCode", Widget: openInVSCodeBtn},
			{Text: "Remove Project", Widget: removeProjectBtn},
			{Text: "README Viewer", Widget: readmeScrollContainer},
		},
	}

	formScroll := container.NewScroll(ui.projectDetails)

	ui.projectList.OnSelected = func(id widget.ListItemID) {
		if id < len(ui.currentProjects) {
			ui.selectedProjectIndex = id
			project := ui.currentProjects[id]
			ui.updateProjectDetails(project)
		}
	}

	split := container.NewHSplit(
		projectListContainer,
		formScroll,
	)
	split.Offset = 0.3

	ui.window.SetContent(split)

	ui.loadProjects()
}

// showNewProjectDialog displays a dialog for creating a new project
func (ui *ProjectManagerUI) showNewProjectDialog() {
	pathEntry := widget.NewEntry()
	pathEntry.SetPlaceHolder("Select project directory")

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter project name")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Enter project description (optional)")

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

	content := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Project Path", Widget: container.NewHBox(pathEntry, pathSelectBtn)},
			{Text: "Project Name", Widget: nameEntry},
			{Text: "Description", Widget: descriptionEntry},
		},
	}

	dialog.ShowCustomConfirm("Create New Project", "Create", "Cancel", content, func(b bool) {
		if !b {
			return
		}

		projectPath := pathEntry.Text
		projectName := nameEntry.Text

		if projectPath == "" || projectName == "" {
			dialog.ShowError(fmt.Errorf("project path and name are required"), ui.window)
			return
		}

		err := utils.ValidateProjectPath(projectPath)
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		readmeContent, _ := utils.ReadReadmeFile(projectPath)

		project := &models.Project{
			Name:        projectName,
			Path:        projectPath,
			Description: descriptionEntry.Text,
			LastOpened:  time.Now(),
			ReadmePath:  "",
		}

		if readmeContent != "" {
			readmePath := filepath.Join(projectPath, "README.md")
			err = ioutil.WriteFile(readmePath, []byte(readmeContent), 0644)
			if err == nil {
				project.ReadmePath = readmePath
			}
		}

		metadata := utils.ScanProjectMetadata(projectPath)
		if len(metadata) > 0 {

			project.Tags = []string{}
		}

		err = ui.projectService.CreateProject(project)
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		ui.loadProjects()
	}, ui.window)
}

// showImportProjectsDialog allows selecting directories to import as projects
func (ui *ProjectManagerUI) showImportProjectsDialog() {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}
		if uri == nil {
			return
		}

		basePath := uri.Path()

		projectPaths, err := utils.FindProjectRoots([]string{basePath})
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		confirmImport := dialog.NewConfirm(
			"Import Projects",
			fmt.Sprintf("Found %d potential projects. Import all?", len(projectPaths)),
			func(confirmed bool) {
				if !confirmed {
					return
				}

				var importedProjects []*models.Project
				var errors []error

				for _, path := range projectPaths {
					project := &models.Project{
						Name:       utils.GetProjectName(path),
						Path:       path,
						LastOpened: time.Now(),
					}

					readmeContent, _ := utils.ReadReadmeFile(path)
					if readmeContent != "" {
						readmePath := filepath.Join(path, "README.md")
						err = ioutil.WriteFile(readmePath, []byte(readmeContent), 0644)
						if err == nil {
							project.ReadmePath = readmePath
						}
					}

					metadata := utils.ScanProjectMetadata(path)
					if len(metadata) > 0 {

						project.Tags = extractTagsFromMetadata(metadata)
					}

					err := ui.projectService.CreateProject(project)
					if err != nil {
						errors = append(errors, fmt.Errorf("failed to import %s: %v", path, err))
					} else {
						importedProjects = append(importedProjects, project)
					}
				}

				if len(errors) > 0 {
					errorMsg := "Some projects failed to import:\n"
					for _, e := range errors {
						errorMsg += e.Error() + "\n"
					}
					dialog.ShowError(fmt.Errorf("%s", errorMsg), ui.window)
				}

				ui.loadProjects()
			},
			ui.window,
		)
		confirmImport.Show()
	}, ui.window)
}

// extractTagsFromMetadata identifies project types based on metadata files
func extractTagsFromMetadata(metadata map[string]string) []string {
	var tags []string

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

// updateProjectDetails updates the UI to display the selected project's information
func (ui *ProjectManagerUI) updateProjectDetails(project models.Project) {
	ui.projectDetails.Items[0].Widget.(*widget.Label).SetText(project.Name)

	ui.descriptionEdit.SetText(project.Description)

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

// uploadReadmeFile allows selecting and attaching a README file to the current project
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

		filePath := uc.URI().Path()

		selectedIndex := ui.selectedProjectIndex
		if selectedIndex < 0 || selectedIndex >= len(ui.currentProjects) {
			dialog.ShowError(fmt.Errorf("no project selected"), ui.window)
			return
		}

		project := ui.currentProjects[selectedIndex]
		project.ReadmePath = filePath

		err = ui.projectService.UpdateProject(&project)
		if err != nil {
			dialog.ShowError(err, ui.window)
			return
		}

		ui.updateProjectDetails(project)
	}, ui.window)
}

// loadProjects retrieves and displays projects from the service
func (ui *ProjectManagerUI) loadProjects() {
	projects, err := ui.projectService.ListProjects()
	if err != nil {
		log.Printf("Error loading projects: %v", err)
		return
	}

	ui.currentProjects = projects

	if ui.projectList != nil {
		ui.projectList.Refresh()
	}

	if len(projects) > 0 {
		ui.projectList.Select(0)
	}
}

// performSearch filters projects based on query text
func (ui *ProjectManagerUI) performSearch(query string) {
	if query == "" {
		ui.loadProjects()
		return
	}

	projects, err := ui.projectService.SearchProjects(query)
	if err != nil {
		dialog.ShowError(fmt.Errorf("search failed: %v", err), ui.window)
		return
	}

	ui.currentProjects = projects

	ui.projectList.Refresh()

	ui.selectedProjectIndex = -1
	ui.updateProjectDetails(models.Project{})

	dialog.ShowInformation(
		"Search Results",
		fmt.Sprintf("Found %d projects matching '%s'", len(projects), query),
		ui.window,
	)
}

// Run displays the window and starts the application event loop
func (ui *ProjectManagerUI) Run() {
	ui.window.ShowAndRun()
}
