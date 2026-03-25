package main

import (
	"context"
	"log"
	"os"

	"github.com/musa-labs/paprawatch/api"
	"github.com/musa-labs/paprawatch/watcher"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "paprawatch",
		Usage: "Watch a directory and upload new files to a Papra instance",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "dir",
				Aliases:  []string{"d"},
				Usage:    "Directory to watch for new files",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "url",
				Aliases: []string{"u"},
				Usage:   "Papra instance URL",
				Value:   "https://api.papra.app",
			},
			&cli.StringFlag{
				Name:     "org",
				Aliases:  []string{"o"},
				Usage:    "Papra Organization ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "token",
				Aliases:  []string{"t"},
				Usage:    "Papra API Token",
				Sources:  cli.EnvVars("PAPRA_API_TOKEN"),
				Required: true,
			},
			&cli.StringFlag{
				Name:    "ocr",
				Usage:   "OCR Languages (optional, e.g. 'eng,fra')",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			dir := cmd.String("dir")
			url := cmd.String("url")
			orgID := cmd.String("org")
			token := cmd.String("token")
			ocr := cmd.String("ocr")

			client := api.NewClient(url, orgID, token)

			onFile := func(filePath string) {
				log.Printf("Uploading file: %s", filePath)
				err := client.UploadDocument(filePath, ocr)
				if err != nil {
					log.Printf("Failed to upload %s: %v", filePath, err)
				} else {
					log.Printf("Successfully uploaded %s", filePath)
				}
			}

			w := watcher.NewWatcher(dir, onFile)
			return w.Start()
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
