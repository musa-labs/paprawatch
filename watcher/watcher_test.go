package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatcher(t *testing.T) {
	tmpDir := t.TempDir()

	fileDetected := make(chan string)

	onFile := func(filePath string) {
		fileDetected <- filePath
	}

	w := NewWatcher(tmpDir, onFile)

	go func() {
		err := w.Start()
		if err != nil {
			t.Errorf("Watcher start failed: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	select {
	case detected := <-fileDetected:
		if filepath.Base(detected) != "test.txt" {
			t.Errorf("Expected 'test.txt', got '%s'", filepath.Base(detected))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for file detection")
	}
}
