package cmd

import (
	"fmt"
	"log"

	"MonoclonalSelectionAutomation/report"

	"github.com/spf13/cobra"
)

var reprocessCmd = &cobra.Command{
	Use:   "reprocess [orderID]",
	Short: "重新分析已解压目录中的 .ab1 文件",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		orderID := args[0]
		// ab1Dir := filepath.Join(orderID, "报告成功")

		geneMap, invalidCount, err := report.ScanAb1Files(orderID)
		if err != nil {
			log.Fatalf("❌ 分析失败: %v", err)
		}
		if invalidCount > 0 {
			log.Printf("⚠️ 有 %d 个 ab1 文件命名不规范", invalidCount)
		}
		if err := report.WriteGeneSummaryFiles(orderID, geneMap); err != nil {
			log.Fatalf("❌ 写入汇总失败: %v", err)
		}
		if err := report.WriteHTMLReport(orderID, geneMap, invalidCount); err != nil {
			log.Fatalf("❌ 写入汇总失败: %v", err)
		}
		fmt.Println("✅ 重新分析完成")
	},
}

func init() {
	rootCmd.AddCommand(reprocessCmd)
}
