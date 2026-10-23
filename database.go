package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// backupDatabase performs a backup of SQLite db for given site
func backupDatabase(site Site, toKeep int) {
	// Create backup dir one level up from live DB dir
	backupDir := filepath.Join(filepath.Dir(filepath.Dir(site.DBPath)), "dbBackups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		fmt.Println("Error creating backup directory for site", site.Name, ":", err)
		return
	}

	backupFileName := fmt.Sprintf("%s-%s.db", filepath.Base(site.DBPath), time.Now().Format("2006-01-02-15-04-05"))
	backupPath := filepath.Join(backupDir, backupFileName)

	if err := createDatabaseBackup(site.DBPath, backupPath); err != nil {
		fmt.Println("Error creating db backup for site", site.Name, ":", err)
		return
	}

	fmt.Println("Successfully created db backup for site", site.Name, "at", backupPath)

	cleanupOldBackups(backupDir, filepath.Base(site.DBPath), toKeep)
	syncDbBackupsToS3(backupDir, site)
}

// createDatabaseBackup publishes a complete SQLite snapshot, including committed WAL data.
func createDatabaseBackup(dbPath, backupPath string) error {
	// Never create an empty source database when the configured path is wrong.
	db, err := sqlite3.OpenFlags(dbPath, sqlite3.OPEN_READONLY|sqlite3.OPEN_URI)
	if err != nil {
		return fmt.Errorf("open source database: %w", err)
	}
	defer db.Close()
	if err := db.BusyTimeout(5 * time.Second); err != nil {
		return fmt.Errorf("set database busy timeout: %w", err)
	}

	// Keep incomplete backups out of retention and S3 sync, and preserve any
	// existing backup at backupPath if this attempt fails.
	tmp, err := os.CreateTemp(filepath.Dir(backupPath), ".sqlite-backup-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary backup: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temporary backup file: %w", err)
	}

	backup, err := db.BackupInit("main", tmpPath)
	if err != nil {
		return fmt.Errorf("initialize SQLite backup: %w", err)
	}
	_, stepErr := backup.Step(-1)
	closeErr := backup.Close()
	if stepErr != nil {
		return fmt.Errorf("copy SQLite backup: %w", stepErr)
	}
	if closeErr != nil {
		return fmt.Errorf("finish SQLite backup: %w", closeErr)
	}
	if err := os.Rename(tmpPath, backupPath); err != nil {
		return fmt.Errorf("publish SQLite backup: %w", err)
	}
	return nil
}

// cleanupOldBackups removes old backups, keep only top N most recent
func cleanupOldBackups(backupDir string, dbName string, toKeep int) {
	files, err := os.ReadDir(backupDir)
	if err != nil {
		fmt.Println("Error reading backup directory for db", dbName, ":", err)
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
				fmt.Println("Error deleting old backup file", filePath, ":", err)
			} else {
				fmt.Println("Deleted old backup file:", filePath)
			}
		}
	}
}
