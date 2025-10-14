package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func CheckUserFolder(userID string) ([]string, error) {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	bucketName := "user-profile"

	listUrl := fmt.Sprintf("%s/storage/v1/object/list/%s", supabaseUrl, bucketName)
	body, _ := json.Marshal(map[string]string{"prefix": userID + "/"})

	req, _ := http.NewRequest("POST", listUrl, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to list files: %s", resp.Status)
	}

	var files []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	var fileNames []string
	for _, f := range files {
		if name, ok := f["name"].(string); ok {
			fileNames = append(fileNames, name)
		}
	}

	return fileNames, nil
}

func DeleteUserPictures(userID string) error {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	bucketName := "user-profile"

	fileNames, err := CheckUserFolder(userID)
	if err != nil {
		return err
	}

	for _, name := range fileNames {
		deleteUrl := fmt.Sprintf("%s/storage/v1/object/%s/%s/%s", supabaseUrl, bucketName, userID, name)
		req, _ := http.NewRequest("DELETE", deleteUrl, nil)
		req.Header.Set("Authorization", "Bearer "+supabaseKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
	}

	return nil
}

func UploadToSupabase(fileHeader *multipart.FileHeader, userID string) (string, error) {
	if err := DeleteUserPictures(userID); err != nil {
		fmt.Println("⚠️ Warning: could not delete old pictures:", err)
	}

	supabaseUrl := os.Getenv("SUPABASE_URL") // e.g. https://yourproject.supabase.co
	supabaseKey := os.Getenv("SUPABASE_KEY") // service_role key (NOT anon)
	bucketName := "user-profile"

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	cleanName := sanitizeFileName(fileHeader.Filename)
	filePath := fmt.Sprintf("%s/%s", userID, cleanName)

	// Upload URL
	uploadUrl := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseUrl, bucketName, filePath)

	// Read file contents
	var buf bytes.Buffer
	_, err = io.Copy(&buf, file)
	if err != nil {
		return "", err
	}

	// Build request
	req, err := http.NewRequest("PUT", uploadUrl, &buf)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", fileHeader.Header.Get("Content-Type"))

	// Execute
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("upload failed: %s", resp.Status)
	}

	// Final public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseUrl, bucketName, filePath)

	return publicURL, nil
}

func sanitizeFileName(filename string) string {
	// Trim spaces at start/end
	filename = strings.TrimSpace(filename)

	// Replace spaces with underscores
	filename = strings.ReplaceAll(filename, " ", "_")

	// Remove special characters (keep letters, numbers, dots, underscores, hyphens)
	reg := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	filename = reg.ReplaceAllString(filename, "")

	return filename
}
