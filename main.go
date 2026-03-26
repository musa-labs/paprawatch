package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/musa-labs/paprawatch/api"
	"github.com/musa-labs/paprawatch/config"
	"github.com/musa-labs/paprawatch/db"
	"github.com/musa-labs/paprawatch/scanner"
	"github.com/musa-labs/paprawatch/watcher"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "paprawatch",
		Usage: "Watch directories and upload new PDF files to a Papra instance",
		Commands: []*cli.Command{
			{
				Name:  "setup",
				Usage: "Run the interactive configuration setup",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := config.LoadConfig()
					if err != nil {
						cfg = &config.Config{}
					}
					return cfg.RunSetup()
				},
			},
			{
				Name:  "scan",
				Usage: "Scan user home directory for PDF files and upload them",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfg, err := config.LoadConfig()
					if err != nil {
						return fmt.Errorf("could not load config: %w", err)
					}

					if cfg.Org == "" || cfg.Token == "" {
						fmt.Println("Configuration is missing. Starting setup...")
						if err := cfg.RunSetup(); err != nil {
							return fmt.Errorf("setup failed: %w", err)
						}
					}

					if cfg.URL == "" {
						cfg.URL = "https://api.papra.app"
					}

					// Merge CLI flags (they take precedence)
					dirs := cmd.StringSlice("dir")
					if len(dirs) > 0 {
						cfg.Dirs = dirs
					}

					// Re-check after setup or merging
					if len(cfg.Dirs) == 0 {
						cfg.Dirs = []string{"."}
					}

					database, err := db.InitDB()
					if err != nil {
						return fmt.Errorf("failed to initialize database: %w", err)
					}
					defer database.Close()

					client := api.NewClient(cfg.URL, cfg.Org, cfg.Token)
					return scanner.Scan(cfg.Dirs, client, database, cfg.OCR)
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "dir",
				Aliases: []string{"d"},
				Usage:   "Directories to watch for new files (can be specified multiple times)",
				Sources: cli.EnvVars("PAPRA_WATCH_DIR"),
			},
			&cli.StringFlag{
				Name:    "url",
				Aliases: []string{"u"},
				Usage:   "Papra instance URL",
				Sources: cli.EnvVars("PAPRA_API_URL"),
			},
			&cli.StringFlag{
				Name:    "org",
				Aliases: []string{"o"},
				Usage:   "Papra Organization ID",
				Sources: cli.EnvVars("PAPRA_ORG_ID"),
			},
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "Papra API Token",
				Sources: cli.EnvVars("PAPRA_API_TOKEN"),
			},
			&cli.StringFlag{
				Name:    "ocr",
				Usage:   "OCR Languages (optional, e.g. 'eng,fra')",
				Sources: cli.EnvVars("PAPRA_OCR_LANGUAGES"),
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				log.Printf("Warning: Could not load config file: %v", err)
				cfg = &config.Config{}
			}

			// Merge CLI flags (they take precedence)
			dirs := cmd.StringSlice("dir")
			if len(dirs) > 0 {
				cfg.Dirs = dirs
			}
			if url := cmd.String("url"); url != "" {
				cfg.URL = url
			}
			if org := cmd.String("org"); org != "" {
				cfg.Org = org
			}
			if token := cmd.String("token"); token != "" {
				cfg.Token = token
			}
			if ocr := cmd.String("ocr"); ocr != "" {
				cfg.OCR = ocr
			}

			// Check if we need to run setup
			if cfg.Org == "" || cfg.Token == "" {
				fmt.Println("Configuration is missing. Starting setup...")
				if err := cfg.RunSetup(); err != nil {
					return fmt.Errorf("setup failed: %w", err)
				}
			}

			// Re-check after setup or merging
			if len(cfg.Dirs) == 0 {
				cfg.Dirs = []string{"."}
			}
			if cfg.Org == "" {
				return fmt.Errorf("organization ID is required")
			}
			if cfg.Token == "" {
				return fmt.Errorf("API token is required")
			}
			if cfg.URL == "" {
				cfg.URL = "https://api.papra.app"
			}

			client := api.NewClient(cfg.URL, cfg.Org, cfg.Token)

			database, err := db.InitDB()
			if err != nil {
				return fmt.Errorf("failed to initialize database: %w", err)
			}
			defer database.Close()

			onFile := func(filePath string) {
				if strings.ToLower(filepath.Ext(filePath)) != ".pdf" {
					return
				}

				// Check if we've seen this file before
				hash, err := scanner.HashFile(filePath)
				if err != nil {
					log.Printf("Failed to hash file %s: %v", filePath, err)
					return
				}

				exists, err := database.HasFile(hash)
				if err != nil {
					log.Printf("Failed to check database for %s: %v", filePath, err)
					return
				}

				if exists {
					log.Printf("Skipping already uploaded file: %s", filePath)
					return
				}

				log.Printf("Uploading file: %s", filePath)
				err = client.UploadDocument(filePath, cfg.OCR)
				if err != nil {
					if err == api.ErrDocumentAlreadyExists {
						log.Printf("File already exists on server: %s. Recording in database.", filePath)
					} else {
						log.Printf("Failed to upload %s: %v", filePath, err)
						return
					}
				}

				if err := database.RecordFile(hash, filePath); err != nil {
					log.Printf("Failed to record file in database %s: %v", filePath, err)
				} else {
					log.Printf("Successfully uploaded and recorded %s", filePath)
				}
			}

			w := watcher.NewWatcher(cfg.Dirs, onFile)
			return w.Start()
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
