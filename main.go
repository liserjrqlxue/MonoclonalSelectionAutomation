package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <url>", os.Args[0])
	}
	rawURL := os.Args[1]
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		log.Fatalf("Invalid URL: %v", err)
	}

	query := parsedURL.Query()
	orderID := query.Get("order_id")
	userEmail := query.Get("user_email")
	password := query.Get("password")
	key := query.Get("key")

	if orderID == "" || userEmail == "" || password == "" || key == "" {
		log.Fatalf("Missing required parameters: order_id, user_email, password, key")
	}

	// 构造下载URL
	newQuery := url.Values{}
	newQuery.Set("action", "os_all_file")
	newQuery.Set("order_id", orderID)
	newQuery.Set("user_email", userEmail)
	newQuery.Set("password", password)
	newQuery.Set("key", key)

	downloadURL := fmt.Sprintf("%s://%s%s?%s", parsedURL.Scheme, parsedURL.Host, parsedURL.Path, newQuery.Encode())
	fmt.Println("Downloading from:", downloadURL)

	// 创建目标文件夹
	destDir := filepath.Join(orderID)
	zipPath := filepath.Join(destDir, fmt.Sprintf("%s.os_all_file.zip", orderID))
	extractDir := filepath.Join(destDir, fmt.Sprintf("%s.os_all_file", orderID))

	if err := os.MkdirAll(destDir, 0755); err != nil {
		log.Fatalf("Failed to create dir: %v", err)
	}

	// 下载文件
	err = downloadFile(zipPath, downloadURL)
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}
	fmt.Println("Downloaded to:", zipPath)

	// 解压
	err = unzip(zipPath, extractDir)
	if err != nil {
		log.Fatalf("Unzip failed: %v", err)
	}
	fmt.Println("Unzipped to:", extractDir)
}

// 下载文件
func downloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("file create failed: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// 解压 zip 文件
func unzip(zipPath string, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		// 检查目录遍历漏洞
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		srcFile, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(dstFile, srcFile)

		dstFile.Close()
		srcFile.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
