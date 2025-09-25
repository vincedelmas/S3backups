package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println("Error reading config file:", err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		fmt.Println("Error unmarshalling config file:", err)
	}

	for _, site := range config.Sites {
		fmt.Println("Processing site:", site.Name)
		if site.DataFolderPath != "" {
			backupDataFolder(site, config.DBBackupsToKeep)
		} else if site.DBPath != "" {
			backupDatabase(site, config.DBBackupsToKeep)
		}
		syncImagesToS3(site)
	}
}
