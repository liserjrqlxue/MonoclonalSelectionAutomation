package unzip

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// 解码文件名：尝试 UTF-8，失败则尝试 GBK
func decodeFileName(name string) string {
	// 如果包含非法 UTF-8，尝试用 GBK 解码
	if !isUTF8(name) {
		decoded, err := decodeGBK([]byte(name))
		if err == nil {
			return decoded
		}
	}
	return name
}

func decodeGBK(data []byte) (string, error) {
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func isUTF8(s string) bool {
	for _, r := range s {
		if r == '\uFFFD' {
			return false
		}
	}
	return true
}

var ab1Pattern = regexp.MustCompile(`^(\d{4}EG[A-Z])[-_](\d{3}[a-zA-Z0-9]*)-{1,2}(\d+)\.T7`)

func Unzip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	var invalidCount = 0

	for _, f := range r.File {
		// ✅ 修复路径兼容性
		safeName := strings.ReplaceAll(decodeFileName(f.Name), "\\", "/")
		fpath := filepath.Join(destDir, safeName)

		// 避免目录穿越攻击
		// 防止 zip slip 攻击
		if !strings.HasPrefix(fpath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("非法路径: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// ✅ 提取 baseName，用于 .ab1 分析时
		baseName := filepath.Base(safeName)
		if strings.HasSuffix(safeName, ".ab1") {
			match := ab1Pattern.FindStringSubmatch(baseName)
			if match == nil {
				fmt.Printf("⚠️ 非法文件名: [%s]\n", baseName)
				invalidCount++

			}
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		dstFile, err := os.OpenFile(fpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
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
	log.Printf("非法文件: %d", invalidCount)
	return nil
}
