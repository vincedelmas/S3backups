package main

import (
	"fmt"
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
	// Use the configured directory, or create folderBackups beside the data folder.
	backupDir := site.FolderBackupsPath
	if backupDir == "" {
		backupDir = filepath.Join(filepath.Dir(site.DataFolderPath), "folderBackups")
	}
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Println("Error creating backup directory for site", site.Name, ":", err)
		return
	}

	backupFileName := fmt.Sprintf("%s-%s.tar.gz", filepath.Base(site.DataFolderPath), time.Now().Format("2006-01-02-15-04-05"))
	backupPath := filepath.Join(backupDir, backupFileName)

	// Create tar.gz archive. Use -C to change dir during archiving so archive doesn't contain full path
	cmd := exec.Command("tar", "-czf", backupPath, "-C", site.DataFolderPath, ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("Error creating folder backup for site", site.Name, ":", err)
		fmt.Println("Output:", string(output))
		return
	}

	fmt.Println("Successfully created folder backup for site", site.Name, "at", backupPath)

	cleanupOldFolderBackups(backupDir, filepath.Base(site.DataFolderPath), toKeep)
	syncDbBackupsToS3(backupDir, site)
}

// cleanupOldFolderBackups removes old folder backups, keeps only top N most recent
func cleanupOldFolderBackups(backupDir string, folderName string, toKeep int) {
	files, err := os.ReadDir(backupDir)
	if err != nil {
		fmt.Println("Error reading backup directory for folder", folderName, ":", err)
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
				fmt.Println("Error deleting old backup file", filePath, ":", err)
			} else {
				fmt.Println("Deleted old backup file", filePath)
			}
		}
	}
}
