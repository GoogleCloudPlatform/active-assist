package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	// Get the absolute path of the ticketinterfaces folder
	folderPath := "internal/ticketinterfaces/plugins"
	absFolderPath, err := filepath.Abs(folderPath)
	if err != nil {
		fmt.Printf("Failed to get absolute path of %s: %s\n", folderPath, err)
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

	// Build plugins from Go files in the ticketinterfaces folder
	err = filepath.Walk(absFolderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || filepath.Ext(path) != ".go" || filepath.Base(path) == "baseTicketInterface.go" {
			return nil
		}

		// Build the plugin using go build -buildmode=plugin
		pluginName := filepath.Base(path[:len(path)-len(filepath.Ext(path))]) + ".so"
		pluginPath := filepath.Join(pluginsFolderPath, pluginName)
		cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", pluginPath, path)
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Failed to build plugin %s: %s\n", path, err)
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error during plugin build: %s\n", err)
		os.Exit(1)
	}

}