package utils

import "StepLauncher/internal/Core/Platform"

func OsName() string { return platform.OsName() }
func OsArch() string { return platform.OsArch() }
func OsKey() string  { return platform.OsKey() }

func NativeClassifier() string                   { return platform.NativeClassifier() }
func NativeClassifierFor(os, arch string) string { return platform.NativeClassifierFor(os, arch) }
