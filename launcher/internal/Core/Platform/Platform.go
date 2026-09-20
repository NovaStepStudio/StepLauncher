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

func NativeClassifierFor(osName, arch string) string {
	switch osName {
	case "windows":
		switch arch {
		case "x86":
			return "natives-windows-x86"
		case "arm64":
			return "natives-windows-arm64"
		default:
			return "natives-windows"
		}
	case "osx":
		if arch == "arm64" {
			return "natives-macos-arm64"
		}
		return "natives-macos"
	default:
		switch arch {
		case "x86":
			return "natives-linux-i386"
		case "arm64":
			return "natives-linux-arm64"
		default:
			return "natives-linux"
		}
	}
}

func NativeClassifier() string {
	return NativeClassifierFor(OsName(), OsArch())
}

func TotalRAMMB() int64 {
	return totalRAMMB()
}
