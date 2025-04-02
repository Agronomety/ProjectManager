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



## Screenshot of GUI

![screenshot](https://github.com/user-attachments/assets/9e9a28f5-19c1-4ced-acbf-9832489ccd0d)






# Technologies

* Language: Go
* Database: SQLite
* GUI Framework: Fyne


# Prerequisites

* Visual Studio Code (recommended)


## If planning to Clone the Repository:

* Go 1.20+
* Git
* C Compiler (https://www.msys2.org/)



# Installation

## Option 1: Download Pre-built Executables

Pre-built executables are available for Windows, macOS, and Linux in the [releases section](https://github.com/Agronomety/ProjectManager/releases) of the repository.

1. Navigate to the [releases page](https://github.com/Agronomety/ProjectManager/releases)
   
2. Download the appropriate application for your operating system:
   - Windows: `projectmanager-windows.exe`
   - macOS: `projectmanager-macos`
   - Linux: `projectmanager-linux`
     
3. For macOS and Linux users, you may need to make the file executable:
   ```bash
   chmod +x projectmanager-macos  # or projectmanager-linux
   
4. If Windows displays a security warning when launching the application, click 'Run anyway' to proceed.


## Option 2: Download the Repository

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

# Bugs and Errors
If you this recieve this error while trying to run the main.go file:

```bash
imports github.com/go-gl/gl/v2.1/gl: build constraints exclude all Go files in...
```

That means that you don't have a C compiler installed and/or recongised. 
I recommend following the steps in https://www.msys2.org/ to download a C compiler.

And after following all the steps there, as an extra last step I would recommend that
you double check that the PATH in system variables has been updated to include the 
terminal for MSYS2.

To do so press Window + R. Then type in sysdm.cpl and press Enter. 
In Advanced click Enviromnent Variables. Under the System Variables click on PATH and 
then Edit. If you can't see a C:\msys64\ucrt64\bin then click on New and then insert
the line into an empty space and click Ok and then close it down. After that restart
your terminal and now you can run the application.
