package main

// Config holds the config for all sites
type Config struct {
	Sites           []Site `json:"sites"`
	DBBackupsToKeep int    `json:"db_backups_to_keep"`
}

// Site represents a site config
type Site struct {
	Name         string `json:"name"`
	DBPath       string `json:"db_path"`
	ImagesDir    string `json:"images_dir"`
	S3Remote     string `json:"s3_remote"`
	S3Bucket     string `json:"s3_bucket"`
	S3DBPath     string `json:"s3_db_path"`
	S3ImagesPath string `json:"s3_images_path"`
}
