package report

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var pattern = regexp.MustCompile(`^(\d{4}EG[A-Z])-(\d{3}[A-Z0-9]*)-(\d+)\.T7`)

type GeneSummary struct {
	GeneName   string   `json:"gene_name"`
	CloneIDs   []string `json:"clone_ids"`
	CloneCount int      `json:"clone_count"`
}

// 核心函数：扫描某个路径下所有 .ab1，返回合法数据、非法数量
func ScanAb1Files(ab1Root string) (map[string]*GeneSummary, int, error) {
	geneMap := make(map[string]*GeneSummary)
	invalidCount := 0

	err := filepath.WalkDir(ab1Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".ab1") {
			return nil
		}

		filename := d.Name()
		match := pattern.FindStringSubmatch(filename)
		if match == nil {
			log.Printf("⚠️ 文件名不符合命名规则: %s", filename)
			invalidCount++
			return nil
		}

		geneName := match[1] + "_" + match[2]
		cloneID := match[3]

		if _, exists := geneMap[geneName]; !exists {
			geneMap[geneName] = &GeneSummary{
				GeneName: geneName,
			}
		}
		geneMap[geneName].CloneIDs = append(geneMap[geneName].CloneIDs, cloneID)
		geneMap[geneName].CloneCount++
		return nil
	})

	return geneMap, invalidCount, err
}

// 导出 TSV 和 JSON 文件
func WriteGeneSummaryFiles(orderID string, geneMap map[string]*GeneSummary) error {
	outDir := filepath.Join(orderID)
	txtPath := filepath.Join(outDir, "gene_clone_summary.txt")
	jsonPath := filepath.Join(outDir, "gene_clone_summary.json")

	// 写 TXT
	txtFile, err := os.Create(txtPath)
	if err != nil {
		return fmt.Errorf("创建 txt 文件失败: %v", err)
	}
	defer txtFile.Close()

	fmt.Fprintf(txtFile, "GeneName\tCloneCount\tCloneIDs\n")

	var geneNames []string
	for name := range geneMap {
		geneNames = append(geneNames, name)
	}
	sort.Strings(geneNames)

	for _, name := range geneNames {
		summary := geneMap[name]
		fmt.Fprintf(txtFile, "%s\t%d\t%s\n", summary.GeneName, summary.CloneCount, strings.Join(summary.CloneIDs, "、"))
	}

	// 写 JSON
	jsonFile, err := os.Create(jsonPath)
	if err != nil {
		return fmt.Errorf("创建 json 文件失败: %v", err)
	}
	defer jsonFile.Close()

	encoder := json.NewEncoder(jsonFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(geneMap); err != nil {
		return fmt.Errorf("写入 json 失败: %v", err)
	}

	// ➕ 写 rename.txt
	renamePath := filepath.Join(outDir, "rename.txt")
	renameFile, err := os.Create(renamePath)
	if err != nil {
		return fmt.Errorf("无法创建 rename.txt: %v", err)
	}
	defer renameFile.Close()

	var names []string
	for name := range geneMap {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, geneName := range names {
		prefix := strings.Replace(geneName, "_", "-", 1)
		fmt.Fprintf(renameFile, "%s\t%s\n", geneName, prefix)
		fmt.Fprintf(renameFile, "%s0P\t%s\n", geneName, prefix)
	}
	log.Printf("✅ 生成 rename.txt: %s\n", renamePath)

	log.Printf("✅ 生成报告：%s / %s\n", txtPath, jsonPath)
	return nil
}

func WriteHTMLReport(orderID string, geneMap map[string]*GeneSummary, invalidCount int) error {
	htmlPath := filepath.Join(orderID, "gene_clone_summary.html")
	f, err := os.Create(htmlPath)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "<!DOCTYPE html><html><head><meta charset='UTF-8'><title>Gene Clone Summary</title>")
	fmt.Fprintln(f, `<style>table{border-collapse:collapse}th,td{border:1px solid #ccc;padding:4px 8px;text-align:left}</style>`)
	fmt.Fprintln(f, "</head><body>")

	fmt.Fprintf(f, "<h2>Gene Clone Summary</h2>")
	fmt.Fprintf(f, "<p><b>合法 GeneName 个数:</b> %d</p>", len(geneMap))
	fmt.Fprintf(f, "<p><b>命名不合法 ab1 文件数:</b> %d</p>", invalidCount)

	fmt.Fprintln(f, "<table><thead><tr><th>GeneName</th><th>CloneCount</th><th>CloneIDs</th></tr></thead><tbody>")
	var names []string
	for name := range geneMap {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		g := geneMap[name]
		ids := strings.Join(g.CloneIDs, "、")
		fmt.Fprintf(f, "<tr><td>%s</td><td>%d</td><td>%s</td></tr>\n", g.GeneName, g.CloneCount, ids)
	}
	fmt.Fprintln(f, "</tbody></table>")
	fmt.Fprintln(f, "</body></html>")

	fmt.Println("✅ 生成 HTML:", htmlPath)
	return nil
}
