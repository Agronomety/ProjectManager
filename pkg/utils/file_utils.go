package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// FindProjectRoots scans directories for potential project roots
func FindProjectRoots(basePaths []string) ([]string, error) {
	var projectRoots []string

	for _, basePath := range basePaths {
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip non-directories
			if !info.IsDir() {
				return nil
			}

			// Check for common project indicators
			projectIndicators := []string{
				".git",
				"go.mod",
				"package.json",
				"requirements.txt",
				"pom.xml",
				"build.gradle",
			}

			for _, indicator := range projectIndicators {
				indicatorPath := filepath.Join(path, indicator)
				if _, err := os.Stat(indicatorPath); err == nil {
					projectRoots = append(projectRoots, path)
					return filepath.SkipDir // Stop walking this directory
				}
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("error scanning %s: %v", basePath, err)
		}
	}

	return projectRoots, nil
}

// ReadReadmeFile attempts to read README files with various common names
func ReadReadmeFile(projectPath string) (string, error) {
	readmeNames := []string{
		"README.md",
		"readme.md",
		"Readme.md",
		"README.txt",
		"readme.txt",
		"README",
		"readme",
	}

	for _, name := range readmeNames {
		readmePath := filepath.Join(projectPath, name)
		content, err := ioutil.ReadFile(readmePath)
		if err == nil {
			return string(content), nil
		}
	}

	return "", errors.New("no README file found")
}

// ValidateProjectPath checks if a path is a valid project directory
func ValidateProjectPath(path string) error {
	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Must be a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Less strict validation
	// Allow the path if it has at least one file or subdirectory
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf("cannot read directory contents: %s", path)
	}

	if len(entries) == 0 {
		return fmt.Errorf("directory is empty: %s", path)
	}

	// Optional: Soft check for project indicators
	softIndicators := []string{
		".git",
		"go.mod",
		"package.json",
		"src",
		"pkg",
		"README.md",
		"readme.md",
	}

	for _, indicator := range softIndicators {
		if _, err := os.Stat(filepath.Join(path, indicator)); err == nil {
			return nil
		}
	}

	// Warning: Directory seems empty or not a typical project
	log.Printf("Warning: Directory %s might not be a typical project", path)
	return nil
}

// GetProjectName attempts to extract a meaningful project name from the path
func GetProjectName(path string) string {
	// Remove trailing slashes
	path = strings.TrimRight(path, "/\\")

	// Get the base directory name
	return filepath.Base(path)
}

// ScanProjectMetadata collects additional metadata about a project
func ScanProjectMetadata(path string) map[string]string {
	metadata := make(map[string]string)

	// Attempt to read project files for additional info
	files := []string{
		"go.mod",
		"package.json",
		"pyproject.toml",
		"pom.xml",
	}

	for _, filename := range files {
		filePath := filepath.Join(path, filename)
		content, err := ioutil.ReadFile(filePath)
		if err == nil {
			metadata[filename] = string(content)
		}
	}

	return metadata
}
