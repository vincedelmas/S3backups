package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Printf("Error reading config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		fmt.Printf("Error unmarshalling config file: %v", err)
	}

	for _, site := range config.Sites {
		fmt.Printf("Processing site: %s", site.Name)
		if site.DataFolderPath != "" {
			backupDataFolder(site, config.DBBackupsToKeep)
		} else if site.DBPath != "" {
			backupDatabase(site, config.DBBackupsToKeep)
		}
		syncImagesToS3(site)
	}

}
