package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// backupDataFolder backup of whole data folder for given site
func backupDataFolder(site Site, toKeep int) {
	// Create backup dir one level up from data folder
	backupDir := filepath.Join(filepath.Dir(site.DataFolderPath), "folderBackups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Printf("Error creating backup directory for site %s: %v", site.Name, err)
		return
	}

	backupFileName := fmt.Sprintf("%s-%s.tar.gz", filepath.Base(site.DataFolderPath), time.Now().Format("2006-01-02-15-04-05"))
	backupPath := filepath.Join(backupDir, backupFileName)

	// Create tar.gz archive. Use -C to change dir during archiving so archive doesn't contain full path
	cmd := exec.Command("tar", "-czf", backupPath, "-C", site.DataFolderPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		log.Printf("Error creating folder backup for site %s: %v\nOutput: %s", site.Name, err, string(output))
		return
	}

	log.Printf("Successfully created folder backup for site %s at %s", site.Name, backupPath)

	cleanupOldFolderBackups(backupDir, filepath.Base(site.DataFolderPath), toKeep)
	syncDbBackupsToS3(backupDir, site)
}

// cleanupOldFolderBackups removes old folder backups, keeps only top N most recent
func cleanupOldFolderBackups(backupDir string, folderName string, toKeep int) {
	files, err := os.ReadDir(backupDir)
	if err != nil {
		log.Printf("Error reading backup directory for folder %s: %v", folderName, err)
		return
	}

	var backupFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tar.gz") && strings.HasPrefix(file.Name(), folderName) {
			backupFiles = append(backupFiles, file)
		}
	}

	if len(backupFiles) > toKeep {
		sort.Slice(backupFiles, func(i, j int) bool {
			infoI, _ := backupFiles[i].Info()
			infoJ, _ := backupFiles[j].Info()

			return infoI.ModTime().Before(infoJ.ModTime())
		})

		for i := 0; i < len(backupFiles)-toKeep; i++ {
			filePath := filepath.Join(backupDir, backupFiles[i].Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("Error deleting old backup file %s: %v", filePath, err)
			} else {
				log.Printf("Deleted old backup file %s", filePath)
			}
		}
	}
}
