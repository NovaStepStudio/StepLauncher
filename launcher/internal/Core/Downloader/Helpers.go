package downloader

import (
	"fmt"
	"strings"
)

const (
	repoMinecraftLibraries = "https://libraries.minecraft.net"
	repoForgeMaven         = "https://maven.minecraftforge.net"
)

func IsForgeGroup(group string) bool {
	if group == "" {
		return false
	}
	g := strings.ToLower(group)
	return strings.HasPrefix(g, "net.minecraftforge") ||
		strings.HasPrefix(g, "cpw.mods") ||
		strings.HasPrefix(g, "org.modlauncher") ||
		strings.HasPrefix(g, "org.spongepowered") ||
		strings.HasPrefix(g, "io.github.llamalad7") ||
		strings.HasPrefix(g, "org.jline") ||
		strings.HasPrefix(g, "org.ow2.asm")
}

func libraryGroup(lib Library) string {
	if lib.Name == "" {
		return ""
	}
	group := lib.Name
	if i := strings.Index(group, ":"); i > 0 {
		group = group[:i]
	}
	return group
}

func LibraryRepositoryBase(lib Library) string {
	if lib.URL != "" {
		return strings.TrimRight(lib.URL, "/")
	}
	if IsForgeGroup(libraryGroup(lib)) {
		return repoForgeMaven
	}
	return repoMinecraftLibraries
}

func HasRepositoryFallback(lib Library) bool {
	if lib.URL != "" {
		return true
	}
	return IsForgeGroup(libraryGroup(lib))
}

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
