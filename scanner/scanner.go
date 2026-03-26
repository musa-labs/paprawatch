package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/musa-labs/paprawatch/api"
	"github.com/musa-labs/paprawatch/db"
)

func Scan(dirs []string, client *api.Client, database *db.Database, ocr string) error {
	for _, startDir := range dirs {
		log.Printf("Starting scan in: %s", startDir)

		err := filepath.WalkDir(startDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}

			if d.IsDir() {
				name := d.Name()
				if strings.HasPrefix(name, ".") || name == "Library" || name == "Applications" || name == "System" || name == "node_modules" {
					return filepath.SkipDir
				}
				return nil
			}

			if strings.ToLower(filepath.Ext(path)) != ".pdf" {
				return nil
			}

			hash, err := HashFile(path)
			if err != nil {
				log.Printf("Failed to hash file %s: %v", path, err)
				return nil
			}

			exists, err := database.HasFile(hash)
			if err != nil {
				log.Printf("Failed to check database for %s: %v", path, err)
				return nil
			}

			if exists {
				log.Printf("Skipping already uploaded file: %s", path)
				return nil
			}

			log.Printf("Uploading new file: %s", path)
			err = client.UploadDocument(path, ocr)
			if err != nil {
				if err == api.ErrDocumentAlreadyExists {
					log.Printf("File already exists on server: %s. Recording in database.", path)
				} else {
					log.Printf("Failed to upload %s: %v", path, err)
					return nil
				}
			}

			if err := database.RecordFile(hash, path); err != nil {
				log.Printf("Failed to record file in database %s: %v", path, err)
			} else {
				log.Printf("Successfully uploaded and recorded %s", path)
			}

			// Add a small delay between uploads to avoid overwhelming the server
			time.Sleep(100 * time.Millisecond)

			return nil
		})

		if err != nil {
			return fmt.Errorf("error walking directory %s: %w", startDir, err)
		}
	}

	log.Println("Scan complete.")
	return nil
}

func HashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
