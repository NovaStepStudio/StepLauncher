package helpers

import "strconv"

const (
	MinRAM   = 512
	MaxRAM   = 32768
	RAMFrac  = 0.55
	OSResRAM = 512
)

func RecommendedMaxRAM(freeRAM int64) int {
	rec := int(float64(freeRAM) * RAMFrac)
	rec = (rec / 256) * 256
	if rec < MinRAM {
		return MinRAM
	}
	if rec > MaxRAM {
		return MaxRAM
	}
	return rec
}

func RecommendedGCPreset(maxRAM int) string {
	switch {
	case maxRAM >= 6144:
		return "zgc"
	case maxRAM >= 3072:
		return "g1gc_optimized"
	default:
		return "g1gc_basic"
	}
}

func GCFlags(preset string) []string {
	switch preset {
	case "g1gc_basic":
		return []string{"-XX:+UseG1GC", "-XX:MaxGCPauseMillis=50"}
	case "g1gc_optimized":
		return []string{
			"-XX:+UseG1GC",
			"-XX:+UnlockExperimentalVMOptions",
			"-XX:MaxGCPauseMillis=50",
			"-XX:G1HeapRegionSize=16m",
			"-XX:G1NewSizePercent=20",
			"-XX:G1MaxNewSizePercent=60",
			"-XX:G1ReservePercent=20",
			"-XX:G1MixedGCLiveThresholdPercent=85",
			"-XX:G1OldCSetRegionThresholdPercent=5",
			"-XX:+ParallelRefProcEnabled",
		}
	case "zgc":
		return []string{"-XX:+UseZGC", "-XX:+ZGenerational"}
	case "shenandoah":
		return []string{"-XX:+UseShenandoahGC", "-XX:ShenandoahGCMode=iu"}
	default:
		return nil
	}
}

func HWAccelDisableFlags() []string {
	return []string{
		"-Dsun.java2d.d3d=false",
		"-Dsun.java2d.opengl=false",
		"-Dsun.java2d.noddraw=true",
		"-Dsun.java2d.ddoffscreen=false",
	}
}

func GPUEnvVars(preference string) map[string]string {
	switch preference {
	case "dgpu":
		return map[string]string{
			"__NV_PRIME_RENDER_OFFLOAD": "1",
			"__VK_LAYER_NV_optimus":     "NVIDIA_only",
			"__GLX_VENDOR_LIBRARY_NAME": "nvidia",
			"DRI_PRIME":                 "1",
		}
	case "igpu":
		return map[string]string{
			"DRI_PRIME": "0",
		}
	default:
		return nil
	}
}

func CrashReasonLabel(exitCode int) string {
	switch exitCode {
	case 0:
		return "normal_exit"
	case 1:
		return "generic_error"
	case -1:
		return "killed_or_oom"
	case 134:
		return "sigsegv_or_abort"
	case 139:
		return "sigsegv"
	case 143:
		return "sigterm"
	case 130:
		return "sigint"
	default:
		if exitCode > 128 {
			return "signal_" + strconv.Itoa(exitCode-128)
		}
		return "exit_" + strconv.Itoa(exitCode)
	}
}

func CrashCategory(exitCode int, reason string) string {
	if exitCode == 0 {
		return "clean"
	}
	if exitCode == -1 || reason == "killed_or_oom" {
		return "oom_or_killed"
	}
	switch {
	case exitCode == 134 || exitCode == 139 || exitCode == 132 || exitCode == 6:
		return "java_vm_crash"
	case exitCode == 143:
		return "killed"
	case exitCode == 130:
		return "interrupted"
	case exitCode == 1:
		return "game_error"
	case exitCode > 128:
		return "signal"
	default:
		if reason == "generic_error" {
			return "game_error"
		}
		return "unknown"
	}
}
