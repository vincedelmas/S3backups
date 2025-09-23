package main

import (
	"log"
	"os"
	"os/exec"
)

// uploadDbToS3 uploads db backup to S3
func syncDbBackupsToS3(backupDir string, site Site) {
	destination := site.S3Remote + ":" + site.S3Bucket + "/" + site.S3DBPath
	rcloneCmd := exec.Command("rclone", "sync", backupDir, destination, "--create-empty-src-dirs")

	if err := rcloneCmd.Run(); err != nil {
		log.Printf("Error syncing db backups to S3 for site %s: %v", site.Name, err)
	} else {
		log.Printf("Successfully synced db backups to S3 for site %s", site.Name)
	}
}

// syncImagesToS3 synchronizes images dir with S3
func syncImagesToS3(site Site) {
	if site.ImagesDir == "" {
		return
	}

	destination := site.S3Remote + ":" + site.S3Bucket + "/" + site.S3ImagesPath
	rcloneCmd := exec.Command("rclone", "sync", site.ImagesDir, destination,
		"--create-empty-src-dirs", "--transfers=16", "--checkers=16", "--fast-list", "--size-only", "--progress")

	// Pipe rclone output to main log
	rcloneCmd.Stdout = os.Stdout
	rcloneCmd.Stderr = os.Stderr

	if err := rcloneCmd.Run(); err != nil {
		log.Printf("Error syncing images to S3 for site %s: %v", site.Name, err)
	} else {
		log.Printf("Successfully synced images to S3 for site %s", site.Name)
	}
}
