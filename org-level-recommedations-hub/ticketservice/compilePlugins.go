// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.```
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Get the absolute path of the main plugin directory
	pluginDir := "internal/ticketinterfaces/plugins"
	absPluginDir, err := filepath.Abs(pluginDir)
	if err != nil {
		fmt.Printf("Failed to get absolute path of %s: %s\n", pluginDir, err)
		os.Exit(1)
	}

	// Create the plugins folder if it doesn't exist
	pluginsFolderPath := "plugins"
	if _, err := os.Stat(pluginsFolderPath); os.IsNotExist(err) {
		err = os.Mkdir(pluginsFolderPath, 0755)
		if err != nil {
			fmt.Printf("Failed to create plugins folder: %s\n", err)
			os.Exit(1)
		}
	}

	// Traverse the main plugin directory
	err = filepath.Walk(absPluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the main plugin directory itself
		if path == absPluginDir {
			return nil
		}

		// If the entry is a directory, check if it contains any .go files
		if info.IsDir() {
			// if it's a subdirectory, skip it
			if path != filepath.Join(absPluginDir, info.Name()) {
				return filepath.SkipDir
			}
			files, err := filepath.Glob(filepath.Join(path, "*.go"))
			if err != nil {
				return err
			}
			
			if len(files) > 0 {
				// Build the plugin using go build -buildmode=plugin
				pluginName := filepath.Base(path) + ".so"
				pluginPath := filepath.Join(pluginsFolderPath, pluginName)
				cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", pluginPath, path)
				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					fmt.Printf("Failed to build plugin %s: %s\n", path, err)
					fmt.Printf("Stdout: %s\n", stdout.String())
    				fmt.Printf("Stderr: %s\n", stderr.String())
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error during plugin build: %s\n", err)
		os.Exit(1)
	}

	fmt.Println("Plugins successfully built.")
}