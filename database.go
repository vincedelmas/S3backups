package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// backupDatabase performs a backup of SQLite db for given site
func backupDatabase(site Site, toKeep int) {
	db, err := sql.Open("sqlite3", site.DBPath)
	if err != nil {
		fmt.Println("Error opening db for site %s: %v", site.Name, err)
		return
	}
	defer db.Close()

	// WAL checkpoint for consistent backup
	if _, err := db.Exec("PRAGMA wal_checkpoint(FULL);"); err != nil {
		fmt.Println("Error performing WAL checkpoint for site %s: %v", site.Name, err)
		return
	}

	// Create backup dir one level up from live DB dir
	backupDir := filepath.Join(filepath.Dir(filepath.Dir(site.DBPath)), "dbBackups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Println("Error creating backup directory for site %s: %v", site.Name, err)
		return
	}

	backupFileName := fmt.Sprintf("%s-%s.db", filepath.Base(site.DBPath), time.Now().Format("2006-01-02-15-04-05"))
	backupPath := filepath.Join(backupDir, backupFileName)

	// Create copy of db file
	cmd := exec.Command("cp", site.DBPath, backupPath)
	if err := cmd.Run(); err != nil {
		fmt.Println("Error creating db backup for site %s: %v", site.Name, err)
		return
	}

	fmt.Println("Successfully created db backup for site %s at %s", site.Name, backupPath)

	cleanupOldBackups(backupDir, filepath.Base(site.DBPath), toKeep)
	syncDbBackupsToS3(backupDir, site)
}

// cleanupOldBackups removes old backups, keep only top N most recent
func cleanupOldBackups(backupDir string, dbName string, toKeep int) {
	files, err := os.ReadDir(backupDir)
	if err != nil {
		fmt.Println("Error reading backup directory for db %s: %v", dbName, err)
		return
	}

	var backupFiles []os.DirEntry
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".db" && len(file.Name()) > len(dbName) && file.Name()[:len(dbName)] == dbName {
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
				fmt.Println("Error deleting old backup file %s: %v", filePath, err)
			} else {
				fmt.Println("Deleted old backup file %s", filePath)
			}
		}
	}
}
