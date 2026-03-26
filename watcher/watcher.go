package watcher

import (
	"log"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	Dirs   []string
	OnFile func(filePath string)
}

func NewWatcher(dirs []string, onFile func(filePath string)) *Watcher {
	return &Watcher{
		Dirs:   dirs,
		OnFile: onFile,
	}
}

func (w *Watcher) Start() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) {
					log.Printf("New file detected: %s", event.Name)
					// Wait a brief moment for the file to be fully written
					time.Sleep(500 * time.Millisecond)
					w.OnFile(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()

	for _, dir := range w.Dirs {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			return err
		}

		err = watcher.Add(absDir)
		if err != nil {
			log.Printf("Failed to watch directory %s: %v", absDir, err)
			return err
		}
		log.Printf("Watching directory: %s", absDir)
	}

	<-done

	return nil
}
