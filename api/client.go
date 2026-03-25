package api

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
)

type Client struct {
	URL   string
	OrgID string
	Token string
	HTTP  *http.Client
}

func NewClient(url, orgID, token string) *Client {
	return &Client{
		URL:   url,
		OrgID: orgID,
		Token: token,
		HTTP:  &http.Client{},
	}
}

func (c *Client) UploadDocument(filePath string, ocrLanguages string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Detect content type
	buffer := make([]byte, 512)
	n, _ := file.Read(buffer)
	contentType := http.DetectContentType(buffer[:n])

	// Reset file pointer after reading for detection
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("could not reset file pointer: %w", err)
	}

	// Use io.Pipe to stream the multipart data instead of buffering it in memory
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		// Create a custom part header to set the specific Content-Type
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(filePath)))
		h.Set("Content-Type", contentType)

		part, err := writer.CreatePart(h)
		if err != nil {
			return
		}

		if _, err := io.Copy(part, file); err != nil {
			return
		}

		// Add ocrLanguages if provided
		if ocrLanguages != "" {
			if err := writer.WriteField("ocrLanguages", ocrLanguages); err != nil {
				return
			}
		}
	}()

	endpoint := fmt.Sprintf("%s/api/organizations/%s/documents", c.URL, c.OrgID)
	req, err := http.NewRequest("POST", endpoint, pr)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
