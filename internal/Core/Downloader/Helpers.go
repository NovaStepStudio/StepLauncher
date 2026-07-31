package downloader

import "fmt"

func CalcPercent(mbDownloaded, mbTotal float64, filesDone, filesTotal int) float64 {
	if filesTotal > 0 {
		return float64(filesDone) / float64(filesTotal) * 100
	}
	if mbTotal > 0 {
		return (mbDownloaded / mbTotal) * 100
	}
	return 0
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func FormatFailed(tasks []DownloadTask) string {
	if len(tasks) == 0 {
		return ""
	}
	if len(tasks) == 1 {
		return fmt.Sprintf("1 file failed: %s", tasks[0].Dest)
	}
	return fmt.Sprintf("%d files failed (e.g. %s)", len(tasks), tasks[0].Dest)
}