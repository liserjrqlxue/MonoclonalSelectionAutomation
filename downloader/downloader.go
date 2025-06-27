package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type DownloadInfo struct {
	OrderID     string
	DownloadURL string
	OutputDir   string
	ZipPath     string
	ExtractDir  string
}

// 构造新的 URL 和目标路径
func PrepareDownloadURL(rawURL string) (*DownloadInfo, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	query := parsedURL.Query()
	orderID := query.Get("order_id")
	userEmail := query.Get("user_email")
	password := query.Get("password")
	key := query.Get("key")

	if orderID == "" || userEmail == "" || password == "" || key == "" {
		return nil, errors.New("missing required parameters")
	}

	newQuery := url.Values{}
	newQuery.Set("action", "os_all_file")
	newQuery.Set("order_id", orderID)
	newQuery.Set("user_email", userEmail)
	newQuery.Set("password", password)
	newQuery.Set("key", key)

	downloadURL := fmt.Sprintf("%s://%s%s?%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path, newQuery.Encode())
	outputDir := filepath.Join(orderID)
	zipPath := filepath.Join(outputDir, fmt.Sprintf("%s.os_all_file.zip", orderID))
	extractDir := filepath.Join(outputDir, fmt.Sprintf("%s.os_all_file", orderID))

	return &DownloadInfo{
		OrderID:     orderID,
		DownloadURL: downloadURL,
		OutputDir:   outputDir,
		ZipPath:     zipPath,
		ExtractDir:  extractDir,
	}, nil
}

// 带重试的下载函数（单线程）
func DownloadWithRetry(path, url string, retry int) error {
	var err error
	for i := 0; i < retry; i++ {
		err = downloadSingle(path, url)
		if err == nil {
			return nil
		}
		fmt.Printf("Retrying (%d/%d): %v\n", i+1, retry, err)
	}
	return err
}

// 单线程下载
func downloadSingle(path, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad response: %s", resp.Status)
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
