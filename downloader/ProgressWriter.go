package downloader

import (
	"fmt"
	"time"
)

// 实时打印下载进度
type ProgressWriter struct {
	Downloaded int64
	Total      int64
	Start      time.Time
	LastPrint  time.Time
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Downloaded += int64(n)

	// 控制刷新频率，0.2秒打印一次
	now := time.Now()
	if now.Sub(pw.LastPrint) < 200*time.Millisecond {
		return n, nil
	}
	pw.LastPrint = now

	elapsed := now.Sub(pw.Start).Seconds()
	speed := float64(pw.Downloaded) / 1024.0 / elapsed // KB/s

	var etaStr string
	if pw.Total > 0 && speed > 0 {
		remaining := float64(pw.Total-pw.Downloaded) / 1024.0 / speed // 秒
		eta := time.Duration(remaining) * time.Second
		etaStr = eta.Truncate(time.Second).String()
	} else {
		etaStr = "??"
	}

	fmt.Printf("\r%.1f%% (%.2f MB / %.2f MB) [%.1f KB/s] ETA: %s",
		float64(pw.Downloaded)*100/float64(pw.Total),
		float64(pw.Downloaded)/1024.0/1024.0,
		float64(pw.Total)/1024.0/1024.0,
		speed,
		etaStr,
	)
	return n, nil
}
