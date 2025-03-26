package ui

import (
	"fmt"
	"io/ioutil"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/Agronomety/ProjectManager/internal/models"
	"github.com/Agronomety/ProjectManager/internal/service"
	"github.com/Agronomety/ProjectManager/pkg/vscode"
)

type ProjectManagerUI struct {
	app             fyne.App
	window          fyne.Window
	projectService  service.ProjectService
	projectList     *widget.List
	projectDetails  *widget.Form
	descriptionEdit *widget.Entry
	readmeViewer    *widget.Label
	currentProjects []models.Project
	vsCodeLauncher  *vscode.Launcher
}

func NewProjectManagerUI(projectService service.ProjectService) *ProjectManagerUI {
	a := app.New()
	w := a.NewWindow("github.com/Agronomety/ProjectManager - Project Manager")
	w.Resize(fyne.NewSize(1200, 800))

	ui := &ProjectManagerUI{
		app:            a,
		window:         w,
		projectService: projectService,
		vsCodeLauncher: &vscode.Launcher{},
	}

	ui.createUI()
	return ui
}

func (ui *ProjectManagerUI) createUI() {
	// Left side - Project List
	ui.projectList = widget.NewList(
		func() int { return len(ui.currentProjects) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Project Template")
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			label := item.(*widget.Label)
			label.SetText(ui.currentProjects[id].Name)
		},
	)

	// Right side - Project Details
	ui.descriptionEdit = widget.NewMultiLineEntry()
	ui.descriptionEdit.SetPlaceHolder("Enter project description...")

	ui.readmeViewer = widget.NewLabel("No README loaded")
	ui.readmeViewer.Wrapping = fyne.TextWrapWord

	readmeUploadBtn := widget.NewButton("Upload README", ui.uploadReadmeFile)

	// Project Details Form
	ui.projectDetails = &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Project Name", Widget: widget.NewLabel("")},
			{Text: "Description", Widget: ui.descriptionEdit},
			{Text: "README", Widget: readmeUploadBtn},
		},
	}

	// Project List Selection Handler
	ui.projectList.OnSelected = func(id widget.ListItemID) {
		project := ui.currentProjects[id]
		ui.updateProjectDetails(project)
	}

	// Main Layout
	split := container.NewHSplit(
		ui.projectList,
		ui.projectDetails,
	)
	split.Offset = 0.3 // 30% list, 70% details

	ui.window.SetContent(split)
	ui.loadProjects()
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
		selectedIndex := ui.projectList.SelectedIndex()
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
	ui.projectList.Refresh()
}

func (ui *ProjectManagerUI) Run() {
	ui.window.ShowAndRun()
}
