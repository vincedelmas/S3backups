package main

import (
	"encoding/json"
	"log"
	"os"
)

func main() {
	configFile, err := os.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		log.Fatalf("Error unmarshalling config file: %v", err)
	}

	for _, site := range config.Sites {
		log.Printf("Processing site: %s", site.Name)
		backupDatabase(site, config.DBBackupsToKeep)
		syncImagesToS3(site)
	}
}
