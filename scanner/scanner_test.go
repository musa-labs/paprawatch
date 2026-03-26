package scanner

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/musa-labs/paprawatch/api"
	"github.com/musa-labs/paprawatch/db"
)

func TestScan(t *testing.T) {
	// Create a temporary directory for scanning
	tmpDir := t.TempDir()
	pdfFile := filepath.Join(tmpDir, "test.pdf")
	err := os.WriteFile(pdfFile, []byte("fake pdf content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test PDF: %v", err)
	}

	// Create a temporary directory for the DB
	dbDir := t.TempDir()
	database, err := db.InitDBAtPath(dbDir)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer database.Close()

	// Create a mock server
	uploadCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uploadCount++
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"id":"doc123"}`)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, "org123", "test-token")
	
	// First scan - should upload the file
	err = Scan([]string{tmpDir}, client, database, "")
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if uploadCount != 1 {
		t.Errorf("Expected 1 upload, got %d", uploadCount)
	}

	// Second scan - should skip the file
	err = Scan([]string{tmpDir}, client, database, "")
	if err != nil {
		t.Fatalf("Second Scan failed: %v", err)
	}

	if uploadCount != 1 {
		t.Errorf("Expected upload count to still be 1, but got %d", uploadCount)
	}

	// Verify database record
	hash, _ := HashFile(pdfFile)
	exists, err := database.HasFile(hash)
	if err != nil {
		t.Fatalf("DB check failed: %v", err)
	}
	if !exists {
		t.Error("Expected file to be recorded in DB")
	}
}
