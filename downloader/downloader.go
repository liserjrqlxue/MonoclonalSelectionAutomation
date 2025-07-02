package downloader

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type DownloadInfo struct {
	OrderID     string
	DownloadURL string
	OutputDir   string
	ZipPath     string
	ExtractDir  string
}

// 构造新的 URL 和输出路径
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

// 下载（带重试和可选SHA256校验）
func DownloadWithRetry(path, url string, retry int, expectedSHA256 string) error {
	var err error
	for i := 0; i < retry; i++ {
		err = downloadSingleWithProgress(path, url, expectedSHA256)
		if err == nil {
			return nil
		}
		fmt.Printf("Retrying (%d/%d): %v\n", i+1, retry, err)
	}
	return err
}

// 下载带进度条 + 可选校验
func downloadSingleWithProgress(path, urlStr string, expectedSHA256 string) error {
	resp, err := http.Get(urlStr)
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

	fmt.Println("Downloading...")

	hasher := sha256.New()
	progress := &ProgressWriter{
		Total: resp.ContentLength, // 或 info.ContentLength
		Start: time.Now(),
	}

	multiWriter := io.MultiWriter(out, hasher, progress)
	_, err = io.Copy(multiWriter, resp.Body)
	fmt.Println() // 换行

	if err != nil {
		return err
	}

	if expectedSHA256 != "" {
		actual := hex.EncodeToString(hasher.Sum(nil))
		if actual != expectedSHA256 {
			return fmt.Errorf("SHA256 mismatch: expected %s, got %s", expectedSHA256, actual)
		}
		fmt.Println("SHA256 checksum verified.")
	}

	return nil
}
