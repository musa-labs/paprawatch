package api

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUploadDocument(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(filePath, []byte("hello papra"), 0644)
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		expectedPath := "/api/organizations/org123/documents"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "Bearer test-token" {
			t.Errorf("Expected Authorization header 'Bearer test-token', got '%s'", authHeader)
		}

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			t.Fatalf("Failed to parse multipart form: %v", err)
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			t.Fatalf("Failed to get form file: %v", err)
		}
		defer file.Close()

		if handler.Filename != "test.txt" {
			t.Errorf("Expected filename 'test.txt', got %s", handler.Filename)
		}

		content, _ := io.ReadAll(file)
		if string(content) != "hello papra" {
			t.Errorf("Expected content 'hello papra', got %s", string(content))
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"id":"doc123"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "org123", "test-token")
	err = client.UploadDocument(filePath, "")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUploadDocument_WithOCR(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(filePath, []byte("hello papra"), 0644)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(10 << 20)
		ocr := r.FormValue("ocrLanguages")
		if ocr != "eng,fra" {
			t.Errorf("Expected ocrLanguages 'eng,fra', got '%s'", ocr)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "org123", "test-token")
	err := client.UploadDocument(filePath, "eng,fra")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestUploadDocument_Failure(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(filePath, []byte("test"), 0644)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
	}))
	defer server.Close()

	client := NewClient(server.URL, "org123", "test-token")
	err := client.UploadDocument(filePath, "")
	if err == nil {
		t.Error("Expected an error for non-2xx response")
	}
	if !strings.Contains(err.Error(), "status 500") {
		t.Errorf("Expected error to contain 'status 500', got: %v", err)
	}
}
