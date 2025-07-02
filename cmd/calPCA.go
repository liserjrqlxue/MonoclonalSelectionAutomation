package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

var calpcaCmd = &cobra.Command{
	Use:   "calpca-run",
	Short: "批量运行 calPCA",
	Run: func(cmd *cobra.Command, args []string) {
		runCalPCA()
	},
}

var (
	calpcaDir     string
	calpcaBase    string
	calpcaOrderID string
)

func init() {
	calpcaCmd.Flags().StringVarP(&calpcaDir, "dir", "d", "", "主目录路径（必须）")
	calpcaCmd.Flags().StringVar(&calpcaBase, "base", "", "baseName（可省略，默认从 dir 获取）")
	calpcaCmd.Flags().StringVar(&calpcaOrderID, "order", "", "订单 ID（可省略，默认从 dir 获取）")

	calpcaCmd.MarkFlagRequired("dir")
	rootCmd.AddCommand(calpcaCmd)
}

func runCalPCA() {
	if calpcaBase == "" {
		calpcaBase = filepath.Base(calpcaDir)
	}
	if calpcaOrderID == "" {
		calpcaOrderID = filepath.Base(calpcaDir)
	}

	root := filepath.Join(calpcaDir, fmt.Sprintf("%s_P_", calpcaBase))

	entries, err := os.ReadDir(root)
	if err != nil {
		log.Fatalf("读取目录失败: %v", err)
	}

	pattern := regexp.MustCompile(fmt.Sprintf("^%s_P_[A-Z]$", calpcaBase))

	for _, entry := range entries {
		if !entry.IsDir() || !pattern.MatchString(entry.Name()) {
			continue
		}

		subBase := entry.Name()
		subDir := filepath.Join(root, subBase)
		prefix := filepath.Join(subDir, subBase)

		args := []string{
			"-i", prefix + "-自合.xlsx",
			"-io", prefix + "-引物订购单_BOM.xlsx",
			"-r", filepath.Join(calpcaOrderID, calpcaOrderID+".os_all_file", "rename.txt"),
			"-s", filepath.Join(calpcaOrderID, calpcaOrderID+".os_all_file"),
			"-o", filepath.Join(calpcaOrderID, subBase),
		}

		runCmd := exec.Command("calPCA", args...)
		fmt.Printf("▶ calPCA in %s\n\tCMD: \t[%s]\n", subBase, runCmd)
		runCmd.Stdout = os.Stdout
		runCmd.Stderr = os.Stderr

		if err := runCmd.Run(); err != nil {
			log.Printf("❌ calPCA 执行失败: %v", err)
		} else {
			fmt.Println("✅ calPCA 成功")
		}
	}
}
