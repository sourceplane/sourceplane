package parser

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourceplane/sourceplane/internal/models"
	"gopkg.in/yaml.v3"
)

// LoadRepository loads and parses an intent.yaml file
func LoadRepository(path string) (*models.Repository, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read intent.yaml: %w", err)
	}

	var repo models.Repository
	if err := yaml.Unmarshal(data, &repo); err != nil {
		return nil, fmt.Errorf("failed to parse intent.yaml: %w", err)
	}

	return &repo, nil
}

// LoadRepositoryFromDir loads intent.yaml from a directory
func LoadRepositoryFromDir(dir string) (*models.Repository, error) {
	path := filepath.Join(dir, "intent.yaml")
	return LoadRepository(path)
}

// LoadBlueprint loads and parses a blueprint.yaml file
func LoadBlueprint(path string) (*models.Blueprint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read blueprint.yaml: %w", err)
	}

	var blueprint models.Blueprint
	if err := yaml.Unmarshal(data, &blueprint); err != nil {
		return nil, fmt.Errorf("failed to parse blueprint.yaml: %w", err)
	}

	return &blueprint, nil
}

// FindIntentYaml searches for intent.yaml in current and parent directories
func FindIntentYaml() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(dir, "intent.yaml")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("intent.yaml not found in current directory or any parent")
		}
		dir = parent
	}
}
