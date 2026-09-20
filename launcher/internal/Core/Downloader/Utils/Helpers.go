package utils

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
