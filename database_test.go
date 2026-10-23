package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/ncruces/go-sqlite3/driver"
)

func TestCreateDatabaseBackup(t *testing.T) {
	for _, journalMode := range []string{"DELETE", "WAL"} {
		t.Run(journalMode, func(t *testing.T) {
			dir := t.TempDir()
			sourcePath := filepath.Join(dir, "source.db")
			backupPath := filepath.Join(dir, "backup.db")
			source, err := sql.Open("sqlite3", sourcePath)
			if err != nil {
				t.Fatal(err)
			}
			defer source.Close()
			source.SetMaxOpenConns(1)
			for _, query := range []string{
				"PRAGMA journal_mode=" + journalMode,
				"PRAGMA wal_autocheckpoint=0",
				"CREATE TABLE records (value TEXT)",
				"INSERT INTO records VALUES ('committed')",
			} {
				if _, err := source.Exec(query); err != nil {
					t.Fatal(err)
				}
			}
			if journalMode == "WAL" {
				info, err := os.Stat(sourcePath + "-wal")
				if err != nil {
					t.Fatal(err)
				}
				if info.Size() == 0 {
					t.Fatal("expected committed data in the WAL")
				}
			}

			// Keep a writer open: the snapshot must include committed rows only.
			tx, err := source.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := tx.Exec("INSERT INTO records VALUES ('uncommitted')"); err != nil {
				t.Fatal(err)
			}
			if err := createDatabaseBackup(sourcePath, backupPath); err != nil {
				t.Fatal(err)
			}

			// Open only the backup file, independently of the live database's WAL.
			backup, err := sql.Open("sqlite3", backupPath)
			if err != nil {
				t.Fatal(err)
			}
			defer backup.Close()
			var integrity, value string
			var count int
			if err := backup.QueryRow("PRAGMA integrity_check").Scan(&integrity); err != nil {
				t.Fatal(err)
			}
			if integrity != "ok" {
				t.Fatalf("backup integrity: %s", integrity)
			}
			if err := backup.QueryRow("SELECT count(*), min(value) FROM records").Scan(&count, &value); err != nil {
				t.Fatal(err)
			}
			if count != 1 || value != "committed" {
				t.Fatalf("backup contains count=%d, value=%q; expected only the committed row", count, value)
			}
		})
	}
}

func TestCreateDatabaseBackupFailure(t *testing.T) {
	for _, sourceKind := range []string{"missing", "invalid"} {
		t.Run(sourceKind, func(t *testing.T) {
			dir := t.TempDir()
			sourcePath := filepath.Join(dir, "source.db")
			backupPath := filepath.Join(dir, "backup.db")
			if sourceKind == "invalid" {
				if err := os.WriteFile(sourcePath, []byte("not a SQLite database"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.WriteFile(backupPath, []byte("existing backup"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := createDatabaseBackup(sourcePath, backupPath); err == nil {
				t.Fatal("expected an error")
			}
			contents, err := os.ReadFile(backupPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(contents) != "existing backup" {
				t.Fatal("failed backup replaced the existing backup")
			}
			if sourceKind == "missing" {
				if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
					t.Fatalf("missing source was created: %v", err)
				}
			}
			leftovers, err := filepath.Glob(filepath.Join(dir, ".sqlite-backup-*"))
			if err != nil {
				t.Fatal(err)
			}
			if len(leftovers) != 0 {
				t.Fatalf("temporary backup files left behind: %v", leftovers)
			}
		})
	}
}
