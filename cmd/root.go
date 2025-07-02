package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"MonoclonalSelectionAutomation/downloader"
	"MonoclonalSelectionAutomation/unzip"
)

var sha256sum string

func init() {
	rootCmd.Flags().StringVar(&sha256sum, "sha256", "", "Expected SHA256 checksum (optional)")
}

var rootCmd = &cobra.Command{
	Use:   "MonoclonalSelectionAutomation <url>",
	Short: "Download and extract os_all_file by parsing a URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rawURL := args[0]
		info, err := downloader.PrepareDownloadURL(rawURL)
		if err != nil {
			log.Fatalf("URL parsing failed: %v", err)
		}

		if err := os.MkdirAll(info.OutputDir, 0755); err != nil {
			log.Fatalf("mkdir failed: %v", err)
		}

		fmt.Println("Downloading from:", info.DownloadURL)
		err = downloader.DownloadWithRetry(info.ZipPath, info.DownloadURL, 3, sha256sum)
		if err != nil {
			log.Fatalf("Download failed: %v", err)
		}

		fmt.Println("Unzipping...")
		err = unzip.Unzip(info.ZipPath, info.ExtractDir)
		if err != nil {
			log.Fatalf("Unzip failed: %v", err)
		}
		fmt.Println("Done.")
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
