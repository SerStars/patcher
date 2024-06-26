package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Manifest struct {
	Metadata struct {
		Build         string `json:"build"`
		Commit        string `json:"commit"`
		ConfirmUpdate bool   `json:"confirm_update"`
	} `json:"metadata"`
	Hashes map[string]string `json:"hashes"`
}

func setReactNavigationName() {
	logger.Debug("Attempting to rename React Navigation...")
	modulesPath := filepath.Join(directory, "Payload", "Rosiecord.app.app", "assets", "_node_modules", ".pnpm")

	if exists, _ := exists(modulesPath); !exists {
		logger.Debug("React Navigation does not exist, no need to rename it.")
		return
	}

	manifestPath := filepath.Join(directory, "Payload", "Rosiecord.app", "manifest.json")

	if exists, _ := exists(manifestPath); !exists {
		logger.Debug("React Navigation does not exist, no need to rename it.")
		return
	}

	content, err := os.ReadFile(manifestPath)

	if err != nil {
		logger.Errorf("Couldn't find manifest.json inside the IPA payload. %v", err)
		exit()
	}

	manifest := Manifest{}
	if err := json.Unmarshal(content, &manifest); err != nil {
		logger.Errorf("Failed to unmarshal manifest.json. %v", err)
		exit()
	}

	if manifest.Hashes == nil {
		logger.Infof("No hashes found in manifest.json. Skipping React Navigation rename.")
		return
	}

	// Change manifest hash path
	for key := range manifest.Hashes {
		if !strings.Contains(key, "@react-navigation+elements") {
			continue
		}

		value := manifest.Hashes[key]
		split := strings.Split(key, "/")

		for idx, segment := range split {
			if !strings.Contains(segment, "@react-navigation+elements") {
				continue
			}

			split[idx] = "@react-navigation+elements@patched"
		}

		delete(manifest.Hashes, key)

		newKey := strings.Join(split, "/")
		manifest.Hashes[newKey] = value
	}

	content, err = json.Marshal(manifest)

	if err != nil {
		logger.Errorf("Failed to marshal modified manifest structure. %v", err)
		return
	}

	err = os.WriteFile(manifestPath, content, 0644)

	if err != nil {
		logger.Errorf("Failed to write modified manifest.json file. %v", err)
		return
	} else {
		logger.Info("Wrote modified manifest.json file.")
	}

	// Rename node_modules module folder(s)
	files, err := os.ReadDir(modulesPath)

	if err != nil {
		logger.Errorf("Failed to read node_modules directory. Skipping React Navigation rename. %v", err)
		return
	}

	directories := filter(files, func(entry fs.DirEntry) bool {
		return strings.Contains(entry.Name(), "@react-navigation+elements")
	})

	for _, directory := range directories {
		currentName := filepath.Join(modulesPath, directory.Name())
		newName := filepath.Join(modulesPath, "@react-navigation+elements@patched")

		if err := os.Rename(currentName, newName); err != nil {
			logger.Errorf("Failed to rename React Navigation directory: %v %v", directory.Name(), err)
		} else {
			logger.Infof("Renamed React Navigation directory: %v", directory.Name())
		}
	}

	logger.Info("Successfully renamed React Navigation directories.")
}
