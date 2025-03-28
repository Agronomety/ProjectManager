# ProjectManager: Project Management Toolkit


ProjectManager is a lightweight, cross-platform project management application designed to help developers organize, track, and quickly access their software projects. 
Built with Go, ProjectManager provides an intuitive interface for managing project metadata, searching projects, and launching development environments.



# Features

🗂️ Project Management

* Add, update, and delete project entries
* Store project metadata including name, path, description, and tags
* Track last opened timestamp



🔍 Project Discovery

* Automatically scan project directories
* Detect project roots based on common framework indicators
* Extract project metadata from configuration files


💻 IDE Integration

* Quick launch projects in Visual Studio Code
* Cross-platform support (Windows, macOS, Linux)


💾 Persistent Storage

* SQLite-based project database
* Configurable storage locations
* Easy project searching and filtering

# Technologies

* Language: Go
* Database: SQLite
* GUI Framework: Fyne


# Prerequisites

* Go 1.20+
* Git
* Visual Studio Code (recommended)

# Installation

Clone the Repository

```bash
git clone https://github.com/Agronomety/ProjectManager.git
cd ProjectManager
```

Then to download the dependencies
```bash
go mod tidy
```

And to run the application
```bash
cd cmd
go run main.go
```
