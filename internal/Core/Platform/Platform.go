package platform

import "runtime"

func OsName() string {
	switch runtime.GOOS {
	case "windows":
		return "windows"
	case "darwin":
		return "osx"
	default:
		return "linux"
	}
}

func OsArch() string {
	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "arm64"
	case "386":
		return "x86"
	}
	return "x86_64"
}

func OsKey() string {
	switch runtime.GOOS {
	case "windows":
		return "windows-x64"
	case "darwin":
		switch runtime.GOARCH {
		case "arm64":
			return "mac-os-arm64"
		default:
			return "mac-os"
		}
	default:
		return "linux"
	}
}

func TotalRAMMB() int64 {
	return totalRAMMB()
}
