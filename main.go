package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	var siteName string
	flag.Func("site", "Back up only the site with this exact name from config.json", func(name string) error {
		if name == "" {
			return fmt.Errorf("site name must not be empty")
		}
		siteName = name
		return nil
	})
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "Unexpected arguments; use --site NAME to select a site")
		os.Exit(2)
	}

	configFile, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading config file:", err)
		os.Exit(1)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		fmt.Fprintln(os.Stderr, "Error unmarshalling config file:", err)
		os.Exit(1)
	}

	sites := config.Sites
	if siteName != "" {
		sites = nil
		for _, site := range config.Sites {
			if site.Name == siteName {
				sites = append(sites, site)
			}
		}
		if len(sites) != 1 {
			fmt.Fprintf(os.Stderr, "Expected exactly one site named %q in config.json; found %d\n", siteName, len(sites))
			os.Exit(1)
		}
	}

	for _, site := range sites {
		fmt.Println("Processing site:", site.Name)
		if site.DataFolderPath != "" {
			backupDataFolder(site, config.DBBackupsToKeep)
		} else if site.DBPath != "" {
			backupDatabase(site, config.DBBackupsToKeep)
		}
		syncImagesToS3(site)
	}
}
